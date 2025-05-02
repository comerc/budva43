package transform

import (
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
)

// Service предоставляет методы для преобразования и замены текста
type Service struct {
	log *slog.Logger
	//
	// Здесь могут быть зависимости, например, репозитории
}

// New создает новый экземпляр сервиса для работы с текстовыми трансформациями
func New() *Service {
	return &Service{
		log: slog.With("module", "service.transform"),
		//
	}
}

// // GetReplacement возвращает текст для замены или пустую строку, если замена не найдена
// func (s *Service) GetReplacement(settings *entity.ReplaceFragment, text string) string {
// 	if settings.Replacements == nil {
// 		return ""
// 	}
// 	replacement, ok := settings.Replacements[text]
// 	if !ok {
// 		return ""
// 	}
// 	return replacement
// }

// // ReplaceText заменяет все фрагменты текста согласно настройкам
// func (s *Service) ReplaceText(settings *entity.ReplaceFragment, text string) string {
// 	if settings.Replacements == nil {
// 		return text
// 	}

// 	result := text
// 	for from, to := range settings.Replacements {
// 		// Здесь может быть реализован более сложный алгоритм замены,
// 		// но для простоты используем стандартную замену строк
// 		if from != "" {
// 			result = s.replaceAll(result, from, to)
// 		}
// 	}

// 	return result
// }

// replaceAll заменяет все вхождения подстроки в строке
// Используется вместо strings.ReplaceAll для возможности
// реализации более сложной логики замены в будущем
// func (s *Service) replaceAll(text, from, to string) string {
// 	return strings.ReplaceAll(text, from, to)
// }

// ReplaceLinks заменяет ссылки в тексте согласно настройкам
func (s *Service) ReplaceLinks(text string, linkReplacements map[string]string) string {
	if len(linkReplacements) == 0 {
		return text
	}

	result := text
	for from, to := range linkReplacements {
		result = strings.ReplaceAll(result, from, to)
	}

	return result
}

// RemoveUnwantedContent удаляет нежелательный контент из текста
func (s *Service) RemoveUnwantedContent(text string, patterns []string) string {
	if len(patterns) == 0 {
		return text
	}

	result := text
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		result = re.ReplaceAllString(result, "")
	}

	return result
}

// FormatText форматирует текст согласно заданным правилам
func (s *Service) FormatText(text string, rules map[string]string) string {
	if len(rules) == 0 {
		return text
	}

	result := text

	// Пример правила: "bold" -> "**{text}**"
	if boldRule, ok := rules["bold"]; ok {
		boldPattern := regexp.MustCompile(`\*\*(.*?)\*\*`)
		result = boldPattern.ReplaceAllString(result, boldRule)
	}

	// Пример правила: "italic" -> "_{text}_"
	if italicRule, ok := rules["italic"]; ok {
		italicPattern := regexp.MustCompile(`_(.*?)_`)
		result = italicPattern.ReplaceAllString(result, italicRule)
	}

	return result
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

// addSourceSign добавляет подпись источника к тексту
func (s *Service) addSourceSign(formattedText *client.FormattedText, title string) error {
	// TODO: выполнить корректный перенос из budva32
	return nil
}

// addSourceLink добавляет ссылку на источник к тексту
func (s *Service) addSourceLink(formattedText *client.FormattedText, title string, message *client.Message) error {
	// TODO: выполнить корректный перенос из budva32
	return nil
}

// AddSources добавляет подпись и ссылку на источник к тексту
func (s *Service) AddSources(formattedText *client.FormattedText, message *client.Message, dstChatId int64) error {
	if source, ok := config.Engine.Sources[message.ChatId]; ok {
		if slices.Contains(source.Sign.For, dstChatId) {
			return s.addSourceSign(formattedText, source.Sign.Title)
		} else if slices.Contains(source.Link.For, dstChatId) {
			return s.addSourceLink(formattedText, source.Link.Title, message)
		}
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
