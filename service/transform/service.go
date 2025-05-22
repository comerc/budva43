package transform

import (
	"fmt"
	"log/slog"
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
	GetNewMessageId(chatId, tmpMessageId int64) (int64, error)
	GetCopiedMessageIds(fromChatMessageId string) ([]string, error)
}

// messageService определяет методы для работы с сообщениями
type messageService interface {
	GetReplyMarkupData(message *client.Message) ([]byte, bool)
}

// Service предоставляет методы для преобразования и замены текста
type Service struct {
	log *slog.Logger
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
		log: slog.With("module", "service.transform"),
		//
		telegramRepo:   telegramRepo,
		storageService: storageService,
		messageService: messageService,
	}
}

// Transform преобразует содержимое сообщения
// TODO: withSources - переставить в конец
func (s *Service) Transform(formattedText *client.FormattedText, withSources bool, src *client.Message, dstChatId int64) error {
	if err := s.addAutoAnswer(formattedText, src); err != nil {
		return fmt.Errorf("addAutoAnswer: %w", err)
	}
	if err := s.replaceMyselfLinks(formattedText, src.ChatId, dstChatId); err != nil {
		return fmt.Errorf("replaceMyselfLinks: %w", err)
	}
	if err := s.replaceFragments(formattedText, dstChatId); err != nil {
		return fmt.Errorf("replaceFragments: %w", err)
	}
	// if err := s.resetEntities(formattedText, dstChatId); err != nil {
	// 	return fmt.Errorf("resetEntities: %w", err)
	// }
	// TODO: только addSources() нужно ограничивать для первого сообщения в альбоме?
	if withSources {
		if err := s.addSources(formattedText, src, dstChatId); err != nil {
			return fmt.Errorf("addSources: %w", err)
		}
	}
	return nil
}

// replaceMyselfLinks заменяет ссылки на текущего бота в тексте
func (s *Service) replaceMyselfLinks(formattedText *client.FormattedText, srcChatId, dstChatId int64) error {
	data, ok := config.Engine.Destinations[dstChatId]
	if !ok {
		return nil
	}
	if !data.ReplaceMyselfLinks.Run {
		return nil
	}
	s.log.Debug("replaceMyselfLinks", "srcChatId", srcChatId, "dstChatId", dstChatId)
	for _, entity := range formattedText.Entities {
		textUrl, ok := entity.Type.(*client.TextEntityTypeTextUrl)
		if !ok {
			continue
		}
		messageLinkInfo, err := s.telegramRepo.GetClient().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
			Url: textUrl.Url,
		})
		if err != nil {
			s.log.Error("GetMessageLinkInfo", "err", err)
			return err
		}
		src := messageLinkInfo.Message
		if src == nil || srcChatId != src.ChatId {
			continue
		}
		isReplaced := false
		fromChatMessageId := fmt.Sprintf("%d:%d", src.ChatId, src.Id)
		toChatMessageIds, err := s.storageService.GetCopiedMessageIds(fromChatMessageId)
		if err != nil {
			return fmt.Errorf("GetCopiedMessageIds: %w", err)
		}
		s.log.Debug("replaceMyselfLinks", "fromChatMessageId", fromChatMessageId, "toChatMessageIds", toChatMessageIds)
		var tmpMessageId int64 = 0
		for _, toChatMessageId := range toChatMessageIds {
			a := strings.Split(toChatMessageId, ":")
			if util.ConvertToInt[int64](a[1]) == dstChatId {
				tmpMessageId = util.ConvertToInt[int64](a[2])
				break
			}
		}
		if tmpMessageId != 0 {
			newMessageId, err := s.storageService.GetNewMessageId(dstChatId, tmpMessageId)
			if err != nil {
				return fmt.Errorf("GetNewMessageId: %w", err)
			}
			messageLink, err := s.telegramRepo.GetClient().GetMessageLink(&client.GetMessageLinkRequest{
				ChatId:    dstChatId,
				MessageId: newMessageId,
			})
			if err != nil {
				s.log.Error("GetMessageLink", "err", err)
				return err
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
	return nil
}

// replaceFragments заменяет фрагменты текста согласно настройкам
func (s *Service) replaceFragments(formattedText *client.FormattedText, dstChatId int64) error {
	destination, ok := config.Engine.Destinations[dstChatId]
	if !ok {
		return nil
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
	return nil
}

// addAutoAnswer добавляет ответ на сообщение
func (s *Service) addAutoAnswer(formattedText *client.FormattedText, src *client.Message) error {
	source, ok := config.Engine.Sources[src.ChatId]
	if !ok {
		return nil
	}
	if !source.AutoAnswer {
		return nil
	}
	data, ok := s.messageService.GetReplyMarkupData(src)
	if !ok {
		return nil
	}
	answer, err := s.telegramRepo.GetClient().GetCallbackQueryAnswer(
		&client.GetCallbackQueryAnswerRequest{
			ChatId:    src.ChatId,
			MessageId: src.Id,
			Payload:   &client.CallbackQueryPayloadData{Data: data},
		},
	)
	if err != nil {
		return fmt.Errorf("GetCallbackQueryAnswer: %w", err)
	}
	err = s.addText(formattedText, escapeMarkdown(answer.Text))
	if err != nil {
		return fmt.Errorf("addText for answer: %w", err)
	}
	return nil
}

// addSources добавляет подпись и ссылку на источник к тексту
func (s *Service) addSources(formattedText *client.FormattedText, src *client.Message, dstChatId int64) error {
	source, ok := config.Engine.Sources[src.ChatId]
	if !ok {
		return nil
	}
	if slices.Contains(source.Sign.For, dstChatId) {
		var err error
		text := source.Sign.Title
		err = s.addText(formattedText, text)
		if err != nil {
			return fmt.Errorf("addText for sign: %w", err)
		}
	}
	if slices.Contains(source.Link.For, dstChatId) {
		var err error
		messageLink, err := s.telegramRepo.GetClient().GetMessageLink(&client.GetMessageLinkRequest{
			ChatId:    src.ChatId,
			MessageId: src.Id,
			ForAlbum:  src.MediaAlbumId != 0,
			// ForComment: false, // удалено в новой версии go-tdlib
		})
		if err != nil {
			return fmt.Errorf("GetMessageLink: %w", err)
		}
		text := fmt.Sprintf("[%s%s](%s)", "\U0001f517", source.Link.Title, messageLink.Link)
		err = s.addText(formattedText, text)
		if err != nil {
			return fmt.Errorf("addText for link: %w", err)
		}
	}
	return nil
}

// addText добавляет новый текст в конец форматированного текста
func (s *Service) addText(formattedText *client.FormattedText, text string) error {
	parsedText, err := s.telegramRepo.GetClient().ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: text,
		ParseMode: &client.TextParseModeMarkdown{
			Version: 2,
		},
	})
	if err != nil {
		return fmt.Errorf("ParseTextEntities: %w", err)
	}
	offset := int32(util.RuneCountForUTF16(formattedText.Text))
	if offset > 0 {
		formattedText.Text += "\n\n"
		offset = offset + 2
	}
	for _, entity := range parsedText.Entities {
		entity.Offset += offset
	}
	formattedText.Text += parsedText.Text
	formattedText.Entities = append(formattedText.Entities, parsedText.Entities...)
	return nil
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
