package text_transform

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/entity"
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

// GetReplacement возвращает текст для замены или пустую строку, если замена не найдена
func (s *Service) GetReplacement(settings *entity.ReplaceFragment, text string) string {
	if settings.Replacements == nil {
		return ""
	}
	replacement, ok := settings.Replacements[text]
	if !ok {
		return ""
	}
	return replacement
}

// ReplaceText заменяет все фрагменты текста согласно настройкам
func (s *Service) ReplaceText(settings *entity.ReplaceFragment, text string) string {
	if settings.Replacements == nil {
		return text
	}

	result := text
	for from, to := range settings.Replacements {
		// Здесь может быть реализован более сложный алгоритм замены,
		// но для простоты используем стандартную замену строк
		if from != "" {
			result = s.replaceAll(result, from, to)
		}
	}

	return result
}

// replaceAll заменяет все вхождения подстроки в строке
// Используется вместо strings.ReplaceAll для возможности
// реализации более сложной логики замены в будущем
func (s *Service) replaceAll(text, from, to string) string {
	return strings.ReplaceAll(text, from, to)
}

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

// EscapeMarkdown экранирует специальные символы Markdown в тексте
func (s *Service) EscapeMarkdown(text string) string {
	// Экранирование специальных символов Markdown
	specialChars := []string{`\`, `*`, `_`, "`", "[", "]", "(", ")", "{", "}", "#", "+", "-", ".", "!"}
	result := text

	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, `\`+char)
	}

	return result
}

// ReplaceMyselfLinks заменяет ссылки на текущего бота в тексте
func (s *Service) ReplaceMyselfLinks(text *client.FormattedText, srcChatID, dstChatID int64) error {
	if text == nil {
		return nil
	}

	// Получаем настройки замены ссылок
	settings, ok := config.Engine.ReplaceMyselfLinks[dstChatID]
	if !ok {
		return nil
	}

	// Заменяем ссылки на текущего бота
	for i, entity := range text.Entities {
		if entity.Type == nil {
			continue
		}

		// Проверяем, является ли сущность ссылкой
		textLink, ok := entity.Type.(*client.TextEntityTypeTextUrl)
		if !ok {
			continue
		}

		// Заменяем ссылки на текущего бота
		if strings.Contains(textLink.Url, fmt.Sprintf("t.me/c/%d", srcChatID)) {
			// Заменяем ссылку на ссылку с целевым чатом
			textLink.Url = strings.Replace(textLink.Url,
				fmt.Sprintf("t.me/c/%d", srcChatID),
				fmt.Sprintf("t.me/c/%d", dstChatID), 1)
			text.Entities[i].Type = textLink
		} else if settings.DeleteExternal {
			// Удаляем внешние ссылки, если такая настройка включена
			length := entity.Length
			offset := entity.Offset

			// Удаляем текстовую ссылку
			text.Text = text.Text[:offset] + text.Text[offset+length:]

			// Корректируем смещения остальных сущностей
			for j := i + 1; j < len(text.Entities); j++ {
				if text.Entities[j].Offset > offset {
					text.Entities[j].Offset -= length
				}
			}

			// Удаляем текущую сущность
			text.Entities = append(text.Entities[:i], text.Entities[i+1:]...)
			i-- // Уменьшаем счетчик, так как мы удалили текущую сущность
		}
	}

	return nil
}

// ReplaceFragments заменяет фрагменты текста согласно настройкам
func (s *Service) ReplaceFragments(text *client.FormattedText, dstChatID int64) error {
	if text == nil {
		return nil
	}

	// Получаем настройки замены фрагментов
	settings, ok := config.Engine.ReplaceFragments[dstChatID]
	if !ok {
		return nil
	}

	// Заменяем фрагменты текста
	for from, to := range settings.Replacements {
		if from == "" {
			continue
		}

		// Проверяем, что длины строк совпадают для корректной замены
		if len([]rune(from)) != len([]rune(to)) {
			s.log.Warn("Длина исходного и заменяемого текста не совпадает, пропускаем замену",
				"from", from,
				"to", to)
			continue
		}

		// Заменяем все вхождения фрагмента
		text.Text = strings.ReplaceAll(text.Text, from, to)
	}

	return nil
}

// AddSourceSign добавляет подпись источника к тексту
func (s *Service) AddSourceSign(text *client.FormattedText, srcChatID int64, dstChatID int64) error {
	if text == nil {
		return nil
	}

	// Ищем источник в конфигурации
	source, ok := config.Engine.Sources[srcChatID]
	if !ok || source.Sign == nil {
		return nil
	}

	// Проверяем, нужно ли добавлять подпись для этого чата
	needSign := false
	for _, chatID := range source.Sign.For {
		if chatID == dstChatID {
			needSign = true
			break
		}
	}

	if !needSign {
		return nil
	}

	// Добавляем подпись к тексту
	if text.Text != "" {
		text.Text += "\n\n"
	}
	text.Text += source.Sign.Title

	return nil
}

// AddSourceLink добавляет ссылку на источник к тексту
func (s *Service) AddSourceLink(text *client.FormattedText, srcChatID int64, dstChatID int64, messageID int64) error {
	if text == nil {
		return nil
	}

	// Ищем источник в конфигурации
	source, ok := config.Engine.Sources[srcChatID]
	if !ok || source.Link == nil {
		return nil
	}

	// Проверяем, нужно ли добавлять ссылку для этого чата
	needLink := false
	for _, chatID := range source.Link.For {
		if chatID == dstChatID {
			needLink = true
			break
		}
	}

	if !needLink {
		return nil
	}

	// Добавляем ссылку к тексту
	if text.Text != "" {
		text.Text += "\n\n"
	}

	// Добавляем текст ссылки
	linkText := source.Link.Title
	linkStart := len([]rune(text.Text))
	text.Text += linkText

	// Создаем сущность-ссылку
	entity := &client.TextEntity{
		Offset: int32(linkStart),
		Length: int32(len([]rune(linkText))),
		Type: &client.TextEntityTypeTextUrl{
			Url: fmt.Sprintf("https://t.me/c/%d/%d", srcChatID, messageID),
		},
	}

	// Добавляем сущность в текст
	text.Entities = append(text.Entities, entity)

	return nil
}
