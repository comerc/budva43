package transform

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/domain"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	// tdlibClient methods
	GetChat(*client.GetChatRequest) (*client.Chat, error)
	GetMessageLinkInfo(*client.GetMessageLinkInfoRequest) (*client.MessageLinkInfo, error)
	GetMessageLink(*client.GetMessageLinkRequest) (*client.MessageLink, error)
	GetCallbackQueryAnswer(*client.GetCallbackQueryAnswerRequest) (*client.CallbackQueryAnswer, error)
	ParseTextEntities(*client.ParseTextEntitiesRequest) (*client.FormattedText, error)
}

//go:generate mockery --name=storageService --exported
type storageService interface {
	GetNewMessageId(chatId, tmpMessageId int64) int64
	GetCopiedMessageIds(chatId, messageId int64) []string
}

//go:generate mockery --name=messageService --exported
type messageService interface {
	GetReplyMarkupData(message *client.Message) []byte
}

// replacement представляет замену entity
type replacement struct {
	OldEntity     *client.TextEntity
	NewText       string                // если пустой - текст не меняем
	NewEntityType client.TextEntityType // если пустой - парсим markdown
}

// Service предоставляет методы для преобразования и замены текста
type Service struct {
	log *log.Logger
	//
	telegramRepo   telegramRepo
	storageService storageService
	messageService messageService
}

// New создает новый экземпляр сервиса для работы с текстовыми трансформациями
func New(
	telegramRepo telegramRepo,
	storageService storageService,
	messageService messageService,
) *Service {
	return &Service{
		log: log.NewLogger(),
		//
		telegramRepo:   telegramRepo,
		storageService: storageService,
		messageService: messageService,
	}
}

// Transform преобразует содержимое сообщения
func (s *Service) Transform(formattedText *client.FormattedText, withSources bool,
	src *client.Message, dstChatId, prevMessageId int64, engineConfig *domain.EngineConfig,
) {
	defer func() {
		s.log.ErrorOrDebug(nil, "",
			"withSources", withSources,
			"srcChatId", src.ChatId,
			"srcId", src.Id,
			"dstChatId", dstChatId,
		)
	}()

	s.addAutoAnswer(formattedText, src, engineConfig)
	s.replaceMyselfLinks(formattedText, src.ChatId, dstChatId, engineConfig)
	s.replaceFragments(formattedText, dstChatId, engineConfig)
	// s.resetEntities(formattedText, dstChatId)
	// TODO: только addSources() нужно ограничивать для первого сообщения в альбоме?
	if withSources {
		s.addSourceSign(formattedText, src, dstChatId, engineConfig)
		s.addSourceLink(formattedText, src, dstChatId, engineConfig)
	}
	if prevMessageId != 0 {
		s.addPrevMessageId(formattedText, src, dstChatId, prevMessageId, engineConfig)
	}
}

// addAutoAnswer добавляет ответ на автоответ
func (s *Service) addAutoAnswer(formattedText *client.FormattedText,
	src *client.Message, engineConfig *domain.EngineConfig,
) {
	var err error
	defer func() {
		s.log.ErrorOrDebug(err, "",
			"srcChatId", src.ChatId,
			"srcId", src.Id,
		)
	}()

	source := engineConfig.Sources[src.ChatId]
	if source == nil {
		err = log.NewError("source not found")
		return
	}
	if !source.AutoAnswer {
		err = log.NewError("source.AutoAnswer is false")
		return
	}

	replyMarkupData := s.messageService.GetReplyMarkupData(src)
	if len(replyMarkupData) == 0 {
		err = log.NewError("replyMarkupData is empty")
		return
	}
	var answer *client.CallbackQueryAnswer
	answer, err = s.telegramRepo.GetCallbackQueryAnswer(
		&client.GetCallbackQueryAnswerRequest{
			ChatId:    src.ChatId,
			MessageId: src.Id,
			Payload:   &client.CallbackQueryPayloadData{Data: replyMarkupData},
		},
	)
	if err != nil {
		return
	}

	s.addText(formattedText, util.EscapeMarkdown(answer.Text))
}

// replaceMyselfLinks заменяет ссылки исходного чата ссылками на копии в целевом чате
// или удаляет ссылки на внешние сообщения
func (s *Service) replaceMyselfLinks(formattedText *client.FormattedText,
	srcChatId, dstChatId int64, engineConfig *domain.EngineConfig,
) {
	var err error
	defer func() {
		s.log.ErrorOrDebug(err, "",
			"srcChatId", srcChatId,
			"dstChatId", dstChatId,
		)
	}()

	destination := engineConfig.Destinations[dstChatId]
	if destination == nil {
		err = log.NewError("destination not found")
		return
	}
	replaceMyselfLinks := destination.ReplaceMyselfLinks
	if replaceMyselfLinks == nil {
		err = log.NewError("replaceMyselfLinks is nil")
		return
	}
	if !replaceMyselfLinks.Run {
		err = log.NewError("replaceMyselfLinks.Run is false")
		return
	}

	var chat *client.Chat
	chat, err = s.telegramRepo.GetChat(&client.GetChatRequest{
		ChatId: srcChatId,
	})
	if err != nil {
		return
	}
	_, isBasicGroup := chat.Type.(*client.ChatTypeBasicGroup)

	// Собираем все замены
	var replacements []*replacement
	for _, entity := range formattedText.Entities {
		switch entityType := entity.Type.(type) {
		case *client.TextEntityTypeUrl:
			if replaceMyselfLinks.Run && !isBasicGroup {
				// Извлекаем URL с учетом UTF-16 смещений
				utf16Text := util.EncodeToUTF16(formattedText.Text)
				utf16URL := utf16Text[entity.Offset : entity.Offset+entity.Length]
				url := util.DecodeFromUTF16(utf16URL)
				src := s.getMessageByLink(url)
				if src == nil || srcChatId != src.ChatId {
					// Не наш чат - не трогаем
					continue
				}
				// Ссылка на тот же исходный чат - пытаемся заменить
				myselfLink, err := s.getMyselfLink(src, dstChatId)
				if err != nil {
					// Не удалось получить ссылку на копию
					if replaceMyselfLinks.DeleteExternal {
						deletedLinkText := replaceMyselfLinks.DeletedLinkText
						if deletedLinkText == "" {
							deletedLinkText = domain.DELETED_LINK // тут не нужен util.EscapeMarkdown()
						}
						// Заменяем для URL без текста
						replacements = append(replacements, &replacement{
							OldEntity: entity,
							NewText:   deletedLinkText,
						})
					}
					continue
				}
				// Успешная замена на ссылку копии
				replacements = append(replacements, &replacement{
					OldEntity:     entity,
					NewText:       myselfLink,
					NewEntityType: &client.TextEntityTypeTextUrl{Url: myselfLink},
				})
			}
		case *client.TextEntityTypeTextUrl:
			if replaceMyselfLinks.Run && !isBasicGroup {
				src := s.getMessageByLink(entityType.Url)
				if src == nil || srcChatId != src.ChatId {
					// Не наш чат - не трогаем
					continue
				}
				// Ссылка на тот же исходный чат - пытаемся заменить
				myselfLink, err := s.getMyselfLink(src, dstChatId)
				if err != nil {
					// Не удалось получить ссылку на копию
					if replaceMyselfLinks.DeleteExternal {
						// Зачёркиваем (текст остаётся, URL удаляется)
						replacements = append(replacements, &replacement{
							OldEntity:     entity,
							NewEntityType: &client.TextEntityTypeStrikethrough{},
						})
					}
					continue
				}
				// Успешная замена URL
				replacements = append(replacements, &replacement{
					OldEntity:     entity,
					NewEntityType: &client.TextEntityTypeTextUrl{Url: myselfLink},
				})
			}
		}
	}

	// Применяем все замены
	s.applyReplacements(formattedText, replacements)
}

// replaceFragments заменяет фрагменты текста
func (s *Service) replaceFragments(formattedText *client.FormattedText,
	dstChatId int64, engineConfig *domain.EngineConfig,
) {
	var err error
	defer func() {
		s.log.ErrorOrDebug(err, "",
			"dstChatId", dstChatId,
		)
	}()

	destination := engineConfig.Destinations[dstChatId]
	if destination == nil {
		err = log.NewError("destination not found")
		return
	}

	for _, replaceFragment := range destination.ReplaceFragments {
		re := regexp.MustCompile("(?i)" + replaceFragment.From)
		if re.FindString(formattedText.Text) != "" {
			// вынесено в engineService.validateConfig()
			// if len(util.EncodeToUTF16(replaceFragment.From)) != len(util.EncodeToUTF16(replaceFragment.To)) {
			// err = log.NewError("длина исходного и заменяемого текста должна быть одинаковой",
			// 	"from", replaceFragment.From,
			// 	"to", replaceFragment.To,
			// )
			// 	return
			// }
			formattedText.Text = re.ReplaceAllString(formattedText.Text, replaceFragment.To)
		}
	}
}

// addSourceSign добавляет подпись источника
func (s *Service) addSourceSign(formattedText *client.FormattedText,
	src *client.Message, dstChatId int64, engineConfig *domain.EngineConfig,
) {
	var err error
	defer func() {
		s.log.ErrorOrDebug(err, "",
			"srcChatId", src.ChatId,
			"srcId", src.Id,
			"dstChatId", dstChatId,
		)
	}()

	source := engineConfig.Sources[src.ChatId]
	if source == nil {
		err = log.NewError("source not found")
		return
	}
	if source.Sign == nil || !slices.Contains(source.Sign.For, dstChatId) {
		err = log.NewError("source.Sign without dstChatId")
		return
	}

	text := source.Sign.Title
	s.addText(formattedText, text)
}

// addSourceLink добавляет ссылку на источник
func (s *Service) addSourceLink(formattedText *client.FormattedText,
	src *client.Message, dstChatId int64, engineConfig *domain.EngineConfig,
) {
	var err error
	defer func() {
		s.log.ErrorOrDebug(err, "",
			"srcChatId", src.ChatId,
			"srcId", src.Id,
			"dstChatId", dstChatId,
		)
	}()

	source := engineConfig.Sources[src.ChatId]
	if source == nil {
		err = log.NewError("source not found")
		return
	}
	if source.Link == nil || !slices.Contains(source.Link.For, dstChatId) {
		err = log.NewError("source.Link without dstChatId")
		return
	}

	var messageLink *client.MessageLink
	messageLink, err = s.telegramRepo.GetMessageLink(&client.GetMessageLinkRequest{
		ChatId:    src.ChatId,
		MessageId: src.Id,
		ForAlbum:  src.MediaAlbumId != 0,
		// ForComment: false, // удалено в новой версии go-tdlib
	})
	if err != nil {
		return
	}

	text := fmt.Sprintf("[%s](%s)", source.Link.Title, messageLink.Link)
	s.addText(formattedText, text)
}

// addPrevMessageId добавляет id предыдущей версии сообщения
func (s *Service) addPrevMessageId(formattedText *client.FormattedText,
	src *client.Message, dstChatId, prevMessageId int64, engineConfig *domain.EngineConfig,
) {
	var err error
	defer func() {
		s.log.ErrorOrDebug(err, "",
			"srcChatId", src.ChatId,
			"srcId", src.Id,
			"dstChatId", dstChatId,
			"prevMessageId", prevMessageId,
		)
	}()

	source := engineConfig.Sources[src.ChatId]
	if source == nil {
		err = log.NewError("source not found")
		return
	}
	prev := source.Prev
	if prev == "" {
		prev = "Prev"
	}

	var messageLink *client.MessageLink
	messageLink, err = s.telegramRepo.GetMessageLink(&client.GetMessageLinkRequest{
		ChatId:    dstChatId,
		MessageId: prevMessageId,
		ForAlbum:  src.MediaAlbumId != 0,
		// ForComment: false, // удалено в новой версии go-tdlib
	})
	if err != nil {
		return
	}

	text := fmt.Sprintf("[%s](%s)", prev, messageLink.Link)
	s.addText(formattedText, text)
}

// addText добавляет текст в formattedText
func (s *Service) addText(formattedText *client.FormattedText, text string) {
	var err error
	defer func() {
		s.log.ErrorOrDebug(err, "",
			"text", text,
		)
	}()

	var parsedText *client.FormattedText
	parsedText, err = s.telegramRepo.ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: text,
		ParseMode: &client.TextParseModeMarkdown{
			Version: 2,
		},
	})
	if err != nil {
		return
	}
	offset := int32(len(util.EncodeToUTF16(formattedText.Text))) //nolint:gosec
	if offset > 0 {
		formattedText.Text += "\n\n"
		offset += 2
	}
	for _, entity := range parsedText.Entities {
		entity.Offset += offset
	}
	formattedText.Text += parsedText.Text
	formattedText.Entities = append(formattedText.Entities, parsedText.Entities...)
}

// applyReplacements применяет замены replacements к formattedText
func (s *Service) applyReplacements(formattedText *client.FormattedText, replacements []*replacement) {
	markdownReplacements := []*replacement{}
	// Сортируем замены по Offset в обратном порядке (от конца к началу)
	// чтобы смещения не сбивались при замене текста
	for i := len(replacements) - 1; i >= 0; i-- {
		replacement := replacements[i]

		if replacement.NewText == "" {
			// Только меняем тип entity, текст не трогаем
			replacement.OldEntity.Type = replacement.NewEntityType
			continue
		}

		oldStart := replacement.OldEntity.Offset
		oldEnd := oldStart + replacement.OldEntity.Length

		// Заменяем текст
		newLength := int32(len(util.EncodeToUTF16(replacement.NewText))) //nolint:gosec
		lengthDelta := newLength - replacement.OldEntity.Length

		// Заменяем текст в formattedText.Text
		formattedText.Text = replaceTextUTF16(
			formattedText.Text, oldStart, oldEnd, replacement.NewText)

		// Обновляем entities
		var newEntities []*client.TextEntity
		for _, entity := range formattedText.Entities {
			if entity == replacement.OldEntity {
				if replacement.NewEntityType == nil {
					// Накапливаем markdown replacements
					markdownReplacements = append(markdownReplacements,
						s.collectMarkdownReplacements(oldStart, replacement.NewText)...)
				} else {
					// Заменяем целевой entity
					newEntities = append(newEntities, &client.TextEntity{
						Offset: oldStart,
						Length: newLength,
						Type:   replacement.NewEntityType,
					})
				}
			} else if entity.Offset >= oldStart && entity.Offset+entity.Length <= oldEnd {
				// Entity полностью внутри заменяемого текста → удаляем
				continue
			} else {
				// Entity не пересекается или находится вне заменяемого текста
				if entity.Offset > oldEnd {
					// Entity после замены → сдвигаем
					entity.Offset += lengthDelta
				}
				newEntities = append(newEntities, entity)
			}
		}
		formattedText.Entities = newEntities
	}
	// Обрабатываем markdown replacements для replacement.NewEntityType == nil
	s.applyMarkdownReplacements(formattedText, markdownReplacements)
}

// getMessageByLink получает сообщение по ссылке
func (s *Service) getMessageByLink(url string) *client.Message {
	var err error
	defer func() {
		s.log.ErrorOrDebug(err, "",
			"url", url,
		)
	}()

	var messageLinkInfo *client.MessageLinkInfo
	messageLinkInfo, err = s.telegramRepo.GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
		Url: url,
	})
	if err != nil {
		return nil
	}
	return messageLinkInfo.Message
}

// getMyselfLink получает ссылку на копию сообщения
func (s *Service) getMyselfLink(src *client.Message, dstChatId int64) (string, error) {
	var err error

	toChatMessageIds := s.storageService.GetCopiedMessageIds(src.ChatId, src.Id)
	var tmpMessageId int64 = 0
	for _, toChatMessageId := range toChatMessageIds {
		a := strings.Split(toChatMessageId, ":")
		if util.ConvertToInt[int64](a[1]) == dstChatId {
			tmpMessageId = util.ConvertToInt[int64](a[2])
			break
		}
	}
	if tmpMessageId == 0 {
		return "", log.NewError("tmpMessageId as 0")
	}
	newMessageId := s.storageService.GetNewMessageId(dstChatId, tmpMessageId)
	if newMessageId == 0 {
		return "", log.NewError("newMessageId as 0")
	}
	var messageLink *client.MessageLink
	messageLink, err = s.telegramRepo.GetMessageLink(&client.GetMessageLinkRequest{
		ChatId:    dstChatId,
		MessageId: newMessageId,
	})
	if err != nil {
		return "", err
	}
	if !messageLink.IsPublic {
		return "", log.NewError("messageLink.IsPublic is false")
	}
	return messageLink.Link, nil
}

// collectMarkdownReplacements собирает замены для markdown текста
func (s *Service) collectMarkdownReplacements(oldStart int32, newText string) []*replacement {
	var err error
	defer func() {
		s.log.ErrorOrDebug(err, "",
			"oldStart", oldStart,
			"newText", newText,
		)
	}()

	var formattedText *client.FormattedText
	formattedText, err = s.telegramRepo.ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: newText,
		ParseMode: &client.TextParseModeMarkdown{
			Version: 2,
		},
	})
	if err != nil {
		return nil
	}

	var markdownReplacements []*replacement

	// Сначала заменяем весь markdown текст на чистый текст
	if formattedText.Text != newText {
		// Создаем replacement для замены markdown текста на чистый
		oldEntity := &client.TextEntity{
			Offset: oldStart,
			Length: int32(len(util.EncodeToUTF16(newText))), //nolint:gosec
			Type:   nil,                                     // тип не важен для замены текста
		}

		markdownReplacements = append(markdownReplacements, &replacement{
			OldEntity:     oldEntity,
			NewText:       formattedText.Text, // заменяем на чистый текст
			NewEntityType: nil,                // это не entity replacement
		})
	}

	// Затем добавляем entities для форматирования
	// ВАЖНО: entities должны быть относительно чистого текста, а не markdown
	for _, entity := range formattedText.Entities {
		newEntity := &client.TextEntity{
			Offset: oldStart + entity.Offset,
			Length: entity.Length,
			Type:   entity.Type,
		}

		markdownReplacements = append(markdownReplacements, &replacement{
			OldEntity:     newEntity,
			NewText:       "", // текст не меняем, только добавляем entity
			NewEntityType: entity.Type,
		})
	}
	return markdownReplacements
}

// applyReplacements применяет замены replacements к formattedText
func (s *Service) applyMarkdownReplacements(formattedText *client.FormattedText, markdownReplacements []*replacement) {
	// Разделяем replacements на замены текста и добавления entities
	var textReplacements []*replacement
	var entityReplacements []*replacement

	for _, replacement := range markdownReplacements {
		if replacement.NewText == "" {
			entityReplacements = append(entityReplacements, replacement)
		} else {
			textReplacements = append(textReplacements, replacement)
		}
	}

	// Сначала применяем замены текста (сортируем по offset в обратном порядке)
	for i := len(textReplacements) - 1; i >= 0; i-- {
		replacement := textReplacements[i]

		oldStart := replacement.OldEntity.Offset
		oldEnd := oldStart + replacement.OldEntity.Length

		// Заменяем текст
		formattedText.Text = replaceTextUTF16(
			formattedText.Text, oldStart, oldEnd, replacement.NewText)

		// Обновляем смещения всех entities после замены
		newLength := int32(len(util.EncodeToUTF16(replacement.NewText))) //nolint:gosec
		lengthDelta := newLength - replacement.OldEntity.Length

		// Обновляем существующие entities
		for _, entity := range formattedText.Entities {
			if entity.Offset > oldEnd {
				entity.Offset += lengthDelta
			}
		}

		// Обновляем entities, которые мы собираемся добавить
		for _, entityReplacement := range entityReplacements {
			if entityReplacement.OldEntity.Offset > oldEnd {
				entityReplacement.OldEntity.Offset += lengthDelta
			}
		}
	}

	// Затем добавляем entities для форматирования
	for _, replacement := range entityReplacements {
		formattedText.Entities = append(formattedText.Entities, replacement.OldEntity)
	}
}

// replaceTextUTF16 заменяет фрагмент текста с учетом UTF-16 смещений
func replaceTextUTF16(text string, startOffset, endOffset int32, newText string) string {
	utf16Text := util.EncodeToUTF16(text)
	utf16NewText := util.EncodeToUTF16(newText)

	newUTF16 := make([]uint16, 0, len(utf16Text)+len(utf16NewText))
	newUTF16 = append(newUTF16, utf16Text[:startOffset]...)
	newUTF16 = append(newUTF16, utf16NewText...)
	newUTF16 = append(newUTF16, utf16Text[endOffset:]...)

	return util.DecodeFromUTF16(newUTF16)
}
