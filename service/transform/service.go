package transform

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/entity"
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
	src *client.Message, dstChatId int64, engineConfig *entity.EngineConfig,
) {
	// Чтобы не дублировать входные параметры в дочерних функциях,
	// хотя это и нарушает атомарность сообщений в логе - компромисс
	defer s.log.ErrorOrDebug(nil, "",
		"withSources", withSources,
		"srcChatId", src.ChatId,
		"srcId", src.Id,
		"dstChatId", dstChatId,
	)

	s.addAutoAnswer(formattedText, src, engineConfig)
	s.replaceMyselfLinks(formattedText, src.ChatId, dstChatId, engineConfig)
	s.replaceFragments(formattedText, dstChatId, engineConfig)
	// s.resetEntities(formattedText, dstChatId)
	// TODO: только addSources() нужно ограничивать для первого сообщения в альбоме?
	if withSources {
		s.addSourceSign(formattedText, src, dstChatId, engineConfig)
		s.addSourceLink(formattedText, src, dstChatId, engineConfig)
	}
}

// replaceMyselfLinks заменяет ссылки исходного чата ссылками на копии в целевом чате
// или удаляет ссылки на внешние сообщения
func (s *Service) replaceMyselfLinks(formattedText *client.FormattedText,
	srcChatId, dstChatId int64, engineConfig *entity.EngineConfig,
) {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

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
	if !replaceMyselfLinks.Run && !replaceMyselfLinks.DeleteExternal {
		err = log.NewError("replaceMyselfLinks is empty")
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

	// Структура для хранения информации о замене
	type replacement struct {
		entityIndex  int
		entity       *client.TextEntity
		newText      string
		newUrl       string
		shouldDelete bool // для зачёркивания ссылки когда не найдено скопированное
	}

	// Собираем все замены
	var replacements []replacement
	for i, entity := range formattedText.Entities {
		switch entityType := entity.Type.(type) {
		case *client.TextEntityTypeUrl:
			if replaceMyselfLinks.Run && !isBasicGroup {
				url := util.DecodeFromUTF16(util.EncodeToUTF16(formattedText.Text[entity.Offset : entity.Offset+entity.Length]))
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
						// Заменяем на "DELETED LINK" (для URL без текста)
						replacements = append(replacements, replacement{
							entityIndex:  i,
							entity:       entity,
							newText:      "DELETED LINK", // TODO: брать значение из replace-myself-link.delete-external (вместо bool)
							shouldDelete: true,
						})
					}
					continue
				}
				// Успешная замена на ссылку копии
				replacements = append(replacements, replacement{
					entityIndex: i,
					entity:      entity,
					newText:     myselfLink,
					newUrl:      myselfLink,
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
						replacements = append(replacements, replacement{
							entityIndex:  i,
							entity:       entity,
							shouldDelete: true,
						})
					}
					continue
				}
				// Успешная замена URL
				replacements = append(replacements, replacement{
					entityIndex: i,
					entity:      entity,
					newUrl:      myselfLink,
				})
			}
		}
	}

	// Применяем замены к тексту в обратном порядке (от конца к началу)
	for i := len(replacements) - 1; i >= 0; i-- {
		replacement := replacements[i]
		if replacement.newText != "" {
			// Заменяем текст ссылки
			start := replacement.entity.Offset
			end := replacement.entity.Offset + replacement.entity.Length
			formattedText.Text = formattedText.Text[:start] + replacement.newText + formattedText.Text[end:]
		}
	}

	// Пересчитываем entities на основе нового текста
	resultEntities := make([]*client.TextEntity, 0, len(formattedText.Entities))
	offsetAdjustment := int32(0)

	for i, entity := range formattedText.Entities {
		// Ищем replacement для этого entity
		var replacement *replacement
		for j := range replacements {
			if replacements[j].entityIndex == i {
				replacement = &replacements[j]
				break
			}
		}

		newEntity := &client.TextEntity{
			Offset: entity.Offset + offsetAdjustment,
			Length: entity.Length,
			Type:   entity.Type,
		}

		// Применяем замену и корректируем смещение
		if replacement != nil {
			if replacement.shouldDelete {
				// Зачёркиваем ссылку
				if replacement.newText != "" {
					newLength := int32(len(util.EncodeToUTF16(replacement.newText)))
					offsetAdjustment += newLength - entity.Length
					newEntity.Length = newLength
				}
				newEntity.Type = &client.TextEntityTypeStrikethrough{}
			} else if replacement.newText != "" {
				// Заменяем на новый текст с URL
				newLength := int32(len(util.EncodeToUTF16(replacement.newText)))
				offsetAdjustment += newLength - entity.Length
				newEntity.Length = newLength
				newEntity.Type = &client.TextEntityTypeTextUrl{
					Url: replacement.newUrl,
				}
			} else if replacement.newUrl != "" {
				// Только меняем URL
				newEntity.Type = &client.TextEntityTypeTextUrl{
					Url: replacement.newUrl,
				}
			}
		}

		resultEntities = append(resultEntities, newEntity)
	}

	formattedText.Entities = resultEntities
}

// getMessageByLink получает сообщение по ссылке - YAGNI (это просто вызов tdlib с логированием)
func (s *Service) getMessageByLink(url string) *client.Message {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

	var messageLinkInfo *client.MessageLinkInfo
	messageLinkInfo, err = s.telegramRepo.GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
		Url: url,
	})
	if err != nil {
		return nil
	}
	return messageLinkInfo.Message
}

// getMyselfLink получает ссылку на копию сообщения в целевом чате
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

// replaceFragments заменяет фрагменты текста согласно настройкам
func (s *Service) replaceFragments(formattedText *client.FormattedText,
	dstChatId int64, engineConfig *entity.EngineConfig,
) {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

	destination := engineConfig.Destinations[dstChatId]
	if destination == nil {
		err = log.NewError("destination not found")
		return
	}

	for _, replaceFragment := range destination.ReplaceFragments {
		re := regexp.MustCompile("(?i)" + replaceFragment.From)
		if re.FindString(formattedText.Text) != "" {
			// вынесено в engineService.validateConfig()
			// if util.RuneCountForUTF16(replaceFragment.From) != util.RuneCountForUTF16(replaceFragment.To) {
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

// addAutoAnswer добавляет ответ на сообщение
func (s *Service) addAutoAnswer(formattedText *client.FormattedText,
	src *client.Message, engineConfig *entity.EngineConfig,
) {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

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

	s.addText(formattedText, escapeMarkdown(answer.Text))
}

func (s *Service) addSourceSign(formattedText *client.FormattedText,
	src *client.Message, dstChatId int64, engineConfig *entity.EngineConfig,
) {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

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

func (s *Service) addSourceLink(formattedText *client.FormattedText,
	src *client.Message, dstChatId int64, engineConfig *entity.EngineConfig,
) {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

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

	text := fmt.Sprintf("[%s%s](%s)", "\U0001f517", source.Link.Title, messageLink.Link)
	s.addText(formattedText, text)
}

// addText добавляет новый текст в конец форматированного текста
func (s *Service) addText(formattedText *client.FormattedText, text string) {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

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
	offset := int32(len(util.EncodeToUTF16(formattedText.Text))) // nolint:gosec
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

// escapeMarkdown экранирует специальные символы Markdown в тексте
func escapeMarkdown(text string) string {
	s := "_ * ( ) ~ ` > # + = | { } . ! \\[ \\] \\-"
	a := strings.Split(s, " ")
	result := text
	for _, v := range a {
		result = strings.ReplaceAll(result, v, "\\"+v)
	}
	return result
	// re := regexp.MustCompile("[" + strings.Join(a, "|") + "]")
	// return re.ReplaceAllString(text, `\$0`)
}
