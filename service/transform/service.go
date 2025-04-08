package text_transform

import (
	"regexp"
	"strings"

	"github.com/comerc/budva43/entity"
)

// TransformService предоставляет методы для преобразования и замены текста
type TransformService struct {
	// Здесь могут быть зависимости, например, репозитории
}

// NewTransformService создает новый экземпляр сервиса для работы с текстовыми трансформациями
func NewTransformService() *TransformService {
	return &TransformService{}
}

// GetReplacement возвращает текст для замены или пустую строку, если замена не найдена
func (s *TransformService) GetReplacement(settings *entity.ReplaceFragmentSettings, text string) string {
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
func (s *TransformService) ReplaceText(settings *entity.ReplaceFragmentSettings, text string) string {
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
func (s *TransformService) replaceAll(text, from, to string) string {
	return strings.ReplaceAll(text, from, to)
}

// ReplaceLinks заменяет ссылки в тексте согласно настройкам
func (s *TransformService) ReplaceLinks(text string, linkReplacements map[string]string) string {
	if linkReplacements == nil || len(linkReplacements) == 0 {
		return text
	}

	result := text
	for from, to := range linkReplacements {
		result = strings.ReplaceAll(result, from, to)
	}

	return result
}

// RemoveUnwantedContent удаляет нежелательный контент из текста
func (s *TransformService) RemoveUnwantedContent(text string, patterns []string) string {
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
func (s *TransformService) FormatText(text string, rules map[string]string) string {
	if rules == nil || len(rules) == 0 {
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
func (s *TransformService) EscapeMarkdown(text string) string {
	// Экранирование специальных символов Markdown
	specialChars := []string{`\`, `*`, `_`, "`", "[", "]", "(", ")", "{", "}", "#", "+", "-", ".", "!"}
	result := text

	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, `\`+char)
	}

	return result
}
