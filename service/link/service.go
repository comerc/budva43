package link

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/comerc/budva43/entity"
)

//go:generate mockery --name=linkProcessor --exported
type linkProcessor interface {
	ProcessLinks(text string) string
}

// LinkService предоставляет методы для работы с ссылками
type LinkService struct {
	processor linkProcessor
}

// NewLinkService создает новый экземпляр сервиса для работы с ссылками
func NewLinkService(processor linkProcessor) *LinkService {
	return &LinkService{processor: processor}
}

// ExtractLinks извлекает все ссылки из текста
func (s *LinkService) ExtractLinks(text string) []string {
	if text == "" {
		return nil
	}

	// Регулярное выражение для поиска URL-адресов
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	return urlRegex.FindAllString(text, -1)
}

// ProcessLinks обрабатывает ссылки в тексте
func (s *LinkService) ProcessLinks(text string) string {
	if text == "" {
		return ""
	}

	if s.processor != nil {
		return s.processor.ProcessLinks(text)
	}

	return text
}

// ReplaceLinks заменяет ссылки в тексте по правилам
func (s *LinkService) ReplaceLinks(text string, replacements map[string]string) string {
	if text == "" || len(replacements) == 0 {
		return text
	}

	result := text
	for from, to := range replacements {
		result = strings.ReplaceAll(result, from, to)
	}

	return result
}

// NormalizeURL нормализует URL-адрес
func (s *LinkService) NormalizeURL(urlStr string) (string, error) {
	if urlStr == "" {
		return "", errors.New("empty URL")
	}

	// Добавляем протокол, если отсутствует
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// Нормализуем путь (убираем дублирующиеся слеши и т.д.)
	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/")

	return parsedURL.String(), nil
}

// ReplaceTelegramLinks заменяет ссылки в тексте на Telegram-ссылки
func (s *LinkService) ReplaceTelegramLinks(text string, chatMapping map[int64]int64) string {
	if text == "" || len(chatMapping) == 0 {
		return text
	}

	// Регулярное выражение для поиска Telegram-ссылок вида t.me/c/1234567890/123
	telegramLinkRegex := regexp.MustCompile(`https?://t\.me/c/(\d+)/(\d+)`)

	result := telegramLinkRegex.ReplaceAllStringFunc(text, func(match string) string {
		parts := telegramLinkRegex.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}

		// Пытаемся конвертировать ID чата
		var chatId int64
		_, err := fmt.Sscanf(parts[1], "%d", &chatId)
		if err != nil {
			return match
		}

		// Если есть маппинг для этого чата, заменяем ID
		if newChatId, ok := chatMapping[chatId]; ok {
			return strings.Replace(match, parts[1], fmt.Sprintf("%d", newChatId), 1)
		}

		return match
	})

	return result
}

// AddTrackingParameters добавляет параметры отслеживания к ссылкам
func (s *LinkService) AddTrackingParameters(text string, params map[string]string) string {
	if text == "" || len(params) == 0 {
		return text
	}

	// Регулярное выражение для поиска URL-адресов
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)

	result := urlRegex.ReplaceAllStringFunc(text, func(match string) string {
		parsedURL, err := url.Parse(match)
		if err != nil {
			return match
		}

		// Получаем текущие параметры запроса
		query := parsedURL.Query()

		// Добавляем новые параметры
		for key, value := range params {
			query.Set(key, value)
		}

		// Обновляем параметры в URL
		parsedURL.RawQuery = query.Encode()

		return parsedURL.String()
	})

	return result
}

// RemoveExternalLinks удаляет внешние ссылки из текста
func (s *LinkService) RemoveExternalLinks(text string, allowedDomains []string) string {
	if text == "" {
		return ""
	}

	// Регулярное выражение для поиска URL-адресов
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)

	result := urlRegex.ReplaceAllStringFunc(text, func(match string) string {
		parsedURL, err := url.Parse(match)
		if err != nil {
			return match
		}

		// Проверяем, входит ли домен в список разрешенных
		for _, domain := range allowedDomains {
			if strings.HasSuffix(parsedURL.Host, domain) {
				return match
			}
		}

		// Если домен не разрешен, удаляем ссылку
		return ""
	})

	return result
}

// GetMessageLink получает ссылку на сообщение
func (s *LinkService) GetMessageLink(message *entity.Message) (string, error) {
	if message == nil {
		return "", errors.New("empty message")
	}

	// Здесь должна быть логика получения ссылки на сообщение
	// с использованием TDLib API или других средств

	// Заглушка
	return fmt.Sprintf("https://t.me/c/%d/%d", message.ChatId, message.Id), nil
}
