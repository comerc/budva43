package transform

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	// tdlibClient methods
	GetMessageLinkInfo(*client.GetMessageLinkInfoRequest) (*client.MessageLinkInfo, error)
	GetMessageLink(*client.GetMessageLinkRequest) (*client.MessageLink, error)
	GetCallbackQueryAnswer(*client.GetCallbackQueryAnswerRequest) (*client.CallbackQueryAnswer, error)
	ParseTextEntities(*client.ParseTextEntitiesRequest) (*client.FormattedText, error)
}

//go:generate mockery --name=storageService --exported
type storageService interface {
	GetNewMessageId(chatId, tmpMessageId int64) int64
	GetCopiedMessageIds(fromChatMessageId string) []string
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
		log: log.NewLogger("service.transform"),
		//
		telegramRepo:   telegramRepo,
		storageService: storageService,
		messageService: messageService,
	}
}

// Transform преобразует содержимое сообщения
// TODO: withSources - переставить в конец
func (s *Service) Transform(formattedText *client.FormattedText, withSources bool, src *client.Message, dstChatId int64) {
	defer s.log.Debug("Transform",
		"withSources", withSources,
		"srcChatId", src.ChatId,
		"srcId", src.Id,
		"dstChatId", dstChatId,
	)

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
	defer s.log.ErrorOrDebug(&err, "replaceMyselfLinks")

	data, ok := config.Engine.Destinations[dstChatId]
	if !ok {
		err = log.NewError("dstChatId not found")
		return
	}
	if !data.ReplaceMyselfLinks.Run {
		err = log.NewError("Run is disabled")
		return
	}
	for _, entity := range formattedText.Entities {
		textUrl, ok := entity.Type.(*client.TextEntityTypeTextUrl)
		if !ok {
			continue
		}
		var messageLinkInfo *client.MessageLinkInfo
		messageLinkInfo, err = s.telegramRepo.GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
			Url: textUrl.Url,
		})
		if err != nil {
			err = log.WrapError(err)
			return
		}
		src := messageLinkInfo.Message
		if src == nil || srcChatId != src.ChatId {
			continue
		}
		isReplaced := false
		fromChatMessageId := fmt.Sprintf("%d:%d", src.ChatId, src.Id)
		toChatMessageIds := s.storageService.GetCopiedMessageIds(fromChatMessageId)
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
				err = log.NewError("GetNewMessageId return 0")
				return
			}
			var messageLink *client.MessageLink
			messageLink, err = s.telegramRepo.GetMessageLink(&client.GetMessageLinkRequest{
				ChatId:    dstChatId,
				MessageId: newMessageId,
			})
			if err != nil {
				err = log.WrapError(err)
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
	defer s.log.ErrorOrDebug(&err, "replaceFragments")

	destination, ok := config.Engine.Destinations[dstChatId]
	if !ok {
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
func (s *Service) addAutoAnswer(formattedText *client.FormattedText, src *client.Message) {
	var err error
	defer s.log.ErrorOrDebug(&err, "addAutoAnswer")

	source, ok := config.Engine.Sources[src.ChatId]
	if !ok {
		err = log.NewError("source not found")
		return
	}
	if !source.AutoAnswer {
		err = log.NewError("source.AutoAnswer is false")
		return
	}
	replyMarkupData := s.messageService.GetReplyMarkupData(src)
	if len(replyMarkupData) == 0 {
		err = log.NewError("reply markup data is empty")
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
		err = log.WrapError(err)
		return
	}
	s.addText(formattedText, escapeMarkdown(answer.Text))
}

// addSources добавляет подпись и ссылку на источник к тексту
func (s *Service) addSources(formattedText *client.FormattedText, src *client.Message, dstChatId int64) {
	var err error
	defer s.log.ErrorOrDebug(&err, "addSources")

	source, ok := config.Engine.Sources[src.ChatId]
	if !ok {
		err = log.NewError("source not found")
		return
	}
	if slices.Contains(source.Sign.For, dstChatId) {
		text := source.Sign.Title
		s.addText(formattedText, text)
	}
	if slices.Contains(source.Link.For, dstChatId) {
		var messageLink *client.MessageLink
		messageLink, err = s.telegramRepo.GetMessageLink(&client.GetMessageLinkRequest{
			ChatId:    src.ChatId,
			MessageId: src.Id,
			ForAlbum:  src.MediaAlbumId != 0,
			// ForComment: false, // удалено в новой версии go-tdlib
		})
		if err != nil {
			err = log.WrapError(err)
			return
		}
		text := fmt.Sprintf("[%s%s](%s)", "\U0001f517", source.Link.Title, messageLink.Link)
		s.addText(formattedText, text)
	}
}

// addText добавляет новый текст в конец форматированного текста
func (s *Service) addText(formattedText *client.FormattedText, text string) {
	var err error
	defer s.log.ErrorOrDebug(&err, "addText")

	var parsedText *client.FormattedText
	parsedText, err = s.telegramRepo.ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: text,
		ParseMode: &client.TextParseModeMarkdown{
			Version: 2,
		},
	})
	if err != nil {
		err = log.WrapError(err)
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
	s1 := "_ * ( ) ~ ` > # + = | { } . !"
	s2 := `\[ \] \-`
	a := strings.Split(s1+" "+s2, " ")
	result := text
	for _, v := range a {
		result = strings.ReplaceAll(result, v, `\`+v)
	}
	return result
	// re := regexp.MustCompile("[" + strings.Join(a, "|") + "]")
	// return re.ReplaceAllString(text, `\$0`)
}
