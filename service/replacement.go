package service

import (
	"github.com/comerc/budva43/entity"
)

// ReplacementService предоставляет методы для работы с заменами текста
type ReplacementService struct {
	// Здесь могут быть зависимости, например, репозитории
}

// NewReplacementService создает новый экземпляр сервиса для работы с заменами текста
func NewReplacementService() *ReplacementService {
	return &ReplacementService{}
}

// GetReplacement возвращает текст для замены или пустую строку, если замена не найдена
func (s *ReplacementService) GetReplacement(settings *entity.ReplaceFragmentSettings, text string) string {
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
func (s *ReplacementService) ReplaceText(settings *entity.ReplaceFragmentSettings, text string) string {
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
func (s *ReplacementService) replaceAll(str, old, new string) string {
	// Пока просто используем наивную замену
	// В будущем можно реализовать более эффективный алгоритм

	// Проверка на пустую строку для замены
	if old == "" {
		return str
	}

	// Проверка на совпадение строк
	if str == old {
		return new
	}

	// Простая замена с использованием временной переменной
	result := ""
	lastIndex := 0
	for i := 0; i <= len(str)-len(old); i++ {
		if str[i:i+len(old)] == old {
			result += str[lastIndex:i] + new
			lastIndex = i + len(old)
			i = lastIndex - 1 // -1 т.к. цикл увеличит i на 1
		}
	}

	// Добавляем оставшуюся часть строки
	if lastIndex < len(str) {
		result += str[lastIndex:]
	}

	return result
}
