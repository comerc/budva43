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

// messageService определяет методы для работы с сообщениями
type messageService interface {
	GetReplyMarkupData(message *client.Message) ([]byte, bool)
}

// Service предоставляет методы для преобразования и замены текста
type Service struct {
	log *slog.Logger
	//
	telegramRepo   telegramRepo
	messageService messageService
}

// New создает новый экземпляр сервиса для работы с текстовыми трансформациями
func New(
	telegramRepo telegramRepo,
	messageService messageService,
) *Service {
	return &Service{
		log: slog.With("module", "service.transform"),
		//
		telegramRepo:   telegramRepo,
		messageService: messageService,
	}
}

// ReplaceMyselfLinks заменяет ссылки на текущего бота в тексте
func (s *Service) ReplaceMyselfLinks(formattedText *client.FormattedText, srcChatId, dstChatId int64) error {
	// TODO: выполнить корректный перенос из budva32
	return nil
}

// ReplaceFragments заменяет фрагменты текста согласно настройкам
func (s *Service) ReplaceFragments(formattedText *client.FormattedText, dstChatId int64) error {
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

// AddSources добавляет подпись и ссылку на источник к тексту
func (s *Service) AddSources(formattedText *client.FormattedText, src *client.Message, dstChatId int64) error {
	source, ok := config.Engine.Sources[src.ChatId]
	if !ok {
		return nil
	}
	if source.AutoAnswer {
		var err error
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
