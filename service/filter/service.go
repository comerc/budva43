package filter

import (
	"log/slog"
	"regexp"
	"strings"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/entity"
)

//go:generate mockery --name=messageService --exported
type messageService interface {
	GetText(message *client.Message) string
	GetCaption(message *client.Message) string
	// GetContentType(message *client.Message) string
}

// Service предоставляет методы для фильтрации сообщений
type Service struct {
	log *slog.Logger
	//
	messageService messageService
}

// New создает новый экземпляр сервиса для фильтрации сообщений
func New(messageService messageService) *Service {
	return &Service{
		log: slog.With("module", "service.filter"),
		//
		messageService: messageService,
	}
}

// MatchesRegexp проверяет, соответствует ли сообщение регулярному выражению
func (s *Service) MatchesRegexp(message *client.Message, pattern string) (bool, error) {
	// Если паттерн пустой, любое сообщение соответствует
	if pattern == "" {
		return true, nil
	}

	// Компилируем регулярное выражение
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	// Получаем текст сообщения или подпись
	text := s.messageService.GetText(message)
	if text == "" {
		text = s.messageService.GetCaption(message)
	}

	// Проверяем соответствие
	return re.MatchString(text), nil
}

// MatchesKeywords проверяет, содержит ли сообщение все ключевые слова
func (s *Service) MatchesKeywords(message *client.Message, keywords []string) bool {
	if len(keywords) == 0 {
		return true
	}

	// Получаем текст сообщения или подпись
	text := s.messageService.GetText(message)
	if text == "" {
		text = s.messageService.GetCaption(message)
	}

	// Приводим текст к нижнему регистру для регистронезависимого поиска
	lowerText := strings.ToLower(text)

	// Проверяем наличие всех ключевых слов
	for _, keyword := range keywords {
		if !strings.Contains(lowerText, strings.ToLower(keyword)) {
			return false
		}
	}

	return true
}

// MatchesAnyKeyword проверяет, содержит ли сообщение хотя бы одно из ключевых слов
func (s *Service) MatchesAnyKeyword(message *client.Message, keywords []string) bool {
	if len(keywords) == 0 {
		return true
	}

	// Получаем текст сообщения или подпись
	text := s.messageService.GetText(message)
	if text == "" {
		text = s.messageService.GetCaption(message)
	}

	// Приводим текст к нижнему регистру для регистронезависимого поиска
	lowerText := strings.ToLower(text)

	// Проверяем наличие хотя бы одного ключевого слова
	for _, keyword := range keywords {
		if strings.Contains(lowerText, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

// MatchesHashtags проверяет, содержит ли сообщение указанные хэштеги
func (s *Service) MatchesHashtags(message *client.Message, hashtags []string) bool {
	if len(hashtags) == 0 {
		return true
	}

	// Получаем текст сообщения или подпись
	text := s.messageService.GetText(message)
	if text == "" {
		text = s.messageService.GetCaption(message)
	}

	// Компилируем регулярное выражение для поиска хэштегов
	re := regexp.MustCompile(`#\w+`)
	foundHashtags := re.FindAllString(text, -1)

	// Проверяем наличие требуемых хэштегов
	for _, requestedTag := range hashtags {
		// Добавляем # к тегу, если его нет
		if !strings.HasPrefix(requestedTag, "#") {
			requestedTag = "#" + requestedTag
		}

		found := false
		for _, foundTag := range foundHashtags {
			if strings.EqualFold(foundTag, requestedTag) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

// // FilterByContentType фильтрует сообщения по типу содержимого
// func (s *Service) FilterByContentType(message *client.Message, allowedTypes []string) bool {
// 	if len(allowedTypes) == 0 {
// 		return true
// 	}

// 	contentType := s.message.GetContentType(message)

// 	for _, allowedType := range allowedTypes {
// 		if contentType == allowedType {
// 			return true
// 		}
// 	}

// 	return false
// }

// ShouldForward проверяет, должно ли сообщение быть переслано согласно правилам
func (s *Service) ShouldForward(message *client.Message, rule *entity.ForwardRule) (bool, error) {
	// Проверка по исключающему регулярному выражению
	if rule.ExcludeRegexp != nil {
		text := s.messageService.GetText(message)
		if text == "" {
			text = s.messageService.GetCaption(message)
		}
		if rule.ExcludeRegexp.MatchString(text) {
			return false, nil
		}
	}

	// Проверка по включающему регулярному выражению
	if rule.IncludeRegexp != nil {
		text := s.messageService.GetText(message)
		if text == "" {
			text = s.messageService.GetCaption(message)
		}
		if !rule.IncludeRegexp.MatchString(text) {
			return false, nil
		}
	}

	// Проверка по подстрокам
	if len(rule.IncludeSubmatch) > 0 {
		text := s.messageService.GetText(message)
		if text == "" {
			text = s.messageService.GetCaption(message)
		}

		matchesAny := false
		for _, submatch := range rule.IncludeSubmatch {
			if submatch.CompiledRegexp != nil {
				matches := submatch.CompiledRegexp.FindStringSubmatch(text)
				if len(matches) > submatch.Group && submatch.Group >= 0 {
					matchValue := matches[submatch.Group]
					for _, allowedMatch := range submatch.Match {
						if matchValue == allowedMatch {
							matchesAny = true
							break
						}
					}
				}
			}
			if matchesAny {
				break
			}
		}

		if !matchesAny {
			return false, nil
		}
	}

	return true, nil
}
