package transform

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/util"
)

// telegramRepo определяет методы для работы с Telegram API
type telegramRepo interface {
	GetClient() *client.Client
}

// storageService определяет методы для работы с хранилищем
type storageService interface {
	GetNewMessageId(chatId, tmpMessageId int64) int64
	GetCopiedMessageIds(fromChatMessageId string) []string
}

// messageService определяет методы для работы с сообщениями
type messageService interface {
	GetReplyMarkupData(message *client.Message) []byte
}

// Service предоставляет методы для преобразования и замены текста
type Service struct {
	log *util.Logger
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
		log: util.NewLogger("service.transform"),
		//
		telegramRepo:   telegramRepo,
		storageService: storageService,
		messageService: messageService,
	}
}

// Transform преобразует содержимое сообщения
// TODO: withSources - переставить в конец
func (s *Service) Transform(formattedText *client.FormattedText, withSources bool, src *client.Message, dstChatId int64) {
	s.addAutoAnswer(formattedText, src)
	s.replaceMyselfLinks(formattedText, src.ChatId, dstChatId)
	s.replaceFragments(formattedText, dstChatId)
	// s.resetEntities(formattedText, dstChatId)
	// TODO: только addSources() нужно ограничивать для первого сообщения в альбоме?
	if withSources {
		s.addSources(formattedText, src, dstChatId)
	}
}

// replaceMyselfLinks заменяет ссылки на текущего бота в тексте
func (s *Service) replaceMyselfLinks(formattedText *client.FormattedText, srcChatId, dstChatId int64) {
	var err error
	defer func() {
		if err != nil {
			// s.log.Error("replaceMyselfLinks", "err", err)
		}
	}()
	data, ok := config.Engine.Destinations[dstChatId]
	if !ok {
		err = fmt.Errorf("destination not found")
		return
	}
	if !data.ReplaceMyselfLinks.Run {
		err = fmt.Errorf("replaceMyselfLinks is disabled")
		return
	}
	// s.log.Debug("replaceMyselfLinks", "srcChatId", srcChatId, "dstChatId", dstChatId)
	for _, entity := range formattedText.Entities {
		textUrl, ok := entity.Type.(*client.TextEntityTypeTextUrl)
		if !ok {
			continue
		}
		var messageLinkInfo *client.MessageLinkInfo
		messageLinkInfo, err = s.telegramRepo.GetClient().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
			Url: textUrl.Url,
		})
		if err != nil {
			// s.log.Error("GetMessageLinkInfo", "err", err)
			return
		}
		src := messageLinkInfo.Message
		if src == nil || srcChatId != src.ChatId {
			continue
		}
		isReplaced := false
		fromChatMessageId := fmt.Sprintf("%d:%d", src.ChatId, src.Id)
		toChatMessageIds := s.storageService.GetCopiedMessageIds(fromChatMessageId)
		// s.log.Debug("replaceMyselfLinks", "fromChatMessageId", fromChatMessageId, "toChatMessageIds", toChatMessageIds)
		var tmpMessageId int64 = 0
		for _, toChatMessageId := range toChatMessageIds {
			a := strings.Split(toChatMessageId, ":")
			if util.ConvertToInt[int64](a[1]) == dstChatId {
				tmpMessageId = util.ConvertToInt[int64](a[2])
				break
			}
		}
		if tmpMessageId != 0 {
			newMessageId := s.storageService.GetNewMessageId(dstChatId, tmpMessageId)
			if newMessageId == 0 {
				err = fmt.Errorf("GetNewMessageId return 0")
				return
			}
			var messageLink *client.MessageLink
			messageLink, err = s.telegramRepo.GetClient().GetMessageLink(&client.GetMessageLinkRequest{
				ChatId:    dstChatId,
				MessageId: newMessageId,
			})
			if err != nil {
				// s.log.Error("GetMessageLink", "err", err)
				return
			}
			entity.Type = &client.TextEntityTypeTextUrl{
				Url: messageLink.Link,
			}
			isReplaced = true
		}
		if !isReplaced && data.ReplaceMyselfLinks.DeleteExternal {
			entity.Type = &client.TextEntityTypeStrikethrough{}
		}
	}
}

// replaceFragments заменяет фрагменты текста согласно настройкам
func (s *Service) replaceFragments(formattedText *client.FormattedText, dstChatId int64) {
	var err error
	defer func() {
		if err != nil {
			// s.log.Error("replaceFragments", "err", err)
		}
	}()
	destination, ok := config.Engine.Destinations[dstChatId]
	if !ok {
		err = fmt.Errorf("destination not found")
		return
	}
	for _, replaceFragment := range destination.ReplaceFragments {
		re := regexp.MustCompile("(?i)" + replaceFragment.From)
		if re.FindString(formattedText.Text) != "" {
			// вынесено в engineService.validateConfig()
			// if util.RuneCountForUTF16(replaceFragment.From) != util.RuneCountForUTF16(replaceFragment.To) {
			// 	return fmt.Errorf("длина исходного и заменяемого текста должна быть одинаковой: %s -> %s", replaceFragment.From, replaceFragment.To)
			// }
			formattedText.Text = re.ReplaceAllString(formattedText.Text, replaceFragment.To)
		}
	}
}

// addAutoAnswer добавляет ответ на сообщение
func (s *Service) addAutoAnswer(formattedText *client.FormattedText, src *client.Message) {
	var err error
	defer func() {
		if err != nil {
			// s.log.Error("addAutoAnswer", "err", err)
		}
	}()
	source, ok := config.Engine.Sources[src.ChatId]
	if !ok {
		err = fmt.Errorf("source not found")
		return
	}
	if !source.AutoAnswer {
		err = fmt.Errorf("source.AutoAnswer is false")
		return
	}
	replyMarkupData := s.messageService.GetReplyMarkupData(src)
	if len(replyMarkupData) == 0 {
		err = fmt.Errorf("reply markup data is empty")
		return
	}
	answer, err := s.telegramRepo.GetClient().GetCallbackQueryAnswer(
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

// addSources добавляет подпись и ссылку на источник к тексту
func (s *Service) addSources(formattedText *client.FormattedText, src *client.Message, dstChatId int64) {
	var err error
	defer func() {
		if err != nil {
			// s.log.Error("addSources", "err", err)
		}
	}()
	source, ok := config.Engine.Sources[src.ChatId]
	if !ok {
		err = fmt.Errorf("source not found")
		return
	}
	if slices.Contains(source.Sign.For, dstChatId) {
		text := source.Sign.Title
		s.addText(formattedText, text)
	}
	if slices.Contains(source.Link.For, dstChatId) {
		// var err error
		messageLink, err := s.telegramRepo.GetClient().GetMessageLink(&client.GetMessageLinkRequest{
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
}

// addText добавляет новый текст в конец форматированного текста
func (s *Service) addText(formattedText *client.FormattedText, text string) {
	var err error
	defer func() {
		if err != nil {
			// TODO: записывать из стека - где вызвана, т.к. вызывается в разных местах
			// s.log.Error("addText", "err", err)
		}
	}()
	parsedText, err := s.telegramRepo.GetClient().ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: text,
		ParseMode: &client.TextParseModeMarkdown{
			Version: 2,
		},
	})
	if err != nil {
		return
	}
	offset := int32(util.RuneCountForUTF16(formattedText.Text)) // nolint:gosec
	if offset > 0 {
		formattedText.Text += "\n\n"
		offset = offset + 2
	}
	for _, entity := range parsedText.Entities {
		entity.Offset += offset
	}
	formattedText.Text += parsedText.Text
	formattedText.Entities = append(formattedText.Entities, parsedText.Entities...)
}

// escapeMarkdown экранирует специальные символы Markdown в тексте
func escapeMarkdown(text string) string {
	// эскейпит все символы: которые нужны для markdown-разметки
	a := []string{
		"_",
		"*",
		`\[`,
		`\]`,
		"(",
		")",
		"~",
		"`",
		">",
		"#",
		"+",
		`\-`,
		"=",
		"|",
		"{",
		"}",
		".",
		"!",
	}
	result := text
	for _, v := range a {
		result = strings.ReplaceAll(result, v, `\`+v)
	}
	return result
	// re := regexp.MustCompile("[" + strings.Join(a, "|") + "]")
	// return re.ReplaceAllString(text, `\$0`)
}
