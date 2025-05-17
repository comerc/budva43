package transform

import (
	"fmt"
	"log/slog"
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

// Service предоставляет методы для преобразования и замены текста
type Service struct {
	log *slog.Logger
	//
	telegramRepo telegramRepo
}

// New создает новый экземпляр сервиса для работы с текстовыми трансформациями
func New(telegramRepo telegramRepo) *Service {
	return &Service{
		log: slog.With("module", "service.transform"),
		//
		telegramRepo: telegramRepo,
	}
}

// ReplaceMyselfLinks заменяет ссылки на текущего бота в тексте
func (s *Service) ReplaceMyselfLinks(formattedText *client.FormattedText, srcChatId, dstChatId int64) error {
	// TODO: выполнить корректный перенос из budva32
	return nil
}

// ReplaceFragments заменяет фрагменты текста согласно настройкам
func (s *Service) ReplaceFragments(formattedText *client.FormattedText, dstChatId int64) error {
	// TODO: выполнить корректный перенос из budva32
	return nil
}

// addSource добавляет источник к тексту
func (s *Service) addSource(formattedText *client.FormattedText, title string) error {
	parsedTitle, err := s.telegramRepo.GetClient().ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: title,
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
	for _, entity := range parsedTitle.Entities {
		entity.Offset += offset
	}
	formattedText.Text += parsedTitle.Text
	formattedText.Entities = append(formattedText.Entities, parsedTitle.Entities...)
	return nil
}

// AddSources добавляет подпись и ссылку на источник к тексту
func (s *Service) AddSources(formattedText *client.FormattedText, message *client.Message, dstChatId int64) error {
	source, ok := config.Engine.Sources[message.ChatId]
	if !ok {
		return nil
	}
	if slices.Contains(source.Sign.For, dstChatId) {
		title := source.Sign.Title
		return s.addSource(formattedText, title)
	}
	if slices.Contains(source.Link.For, dstChatId) {
		messageLink, err := s.telegramRepo.GetClient().GetMessageLink(&client.GetMessageLinkRequest{
			ChatId:    message.ChatId,
			MessageId: message.Id,
			ForAlbum:  message.MediaAlbumId != 0,
			// ForComment: false, // удалено в новой версии go-tdlib
		})
		if err != nil {
			return fmt.Errorf("GetMessageLink: %w", err)
		}
		title := fmt.Sprintf("[%s%s](%s)", "\U0001f517", source.Link.Title, messageLink.Link)
		return s.addSource(formattedText, title)
	}
	return nil
}

// AddAutoAnswer добавляет ответ на сообщение
func (s *Service) AddAutoAnswer(formattedText *client.FormattedText, src *client.Message) {
	// TODO: выполнить корректный перенос из budva32
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
