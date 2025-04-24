package filter

import (
	"log/slog"

	"github.com/comerc/budva43/entity"
)

// Service предоставляет методы для фильтрации сообщений
type Service struct {
	log *slog.Logger
	//
}

// New создает новый экземпляр сервиса для фильтрации сообщений
func New() *Service {
	return &Service{
		log: slog.With("module", "service.filter"),
		//
	}
}

// // MatchesRegexp проверяет, соответствует ли сообщение регулярному выражению
// func (s *Service) MatchesRegexp(text string, pattern string) (bool, error) {
// 	// Если паттерн пустой, любое сообщение соответствует
// 	if pattern == "" {
// 		return true, nil
// 	}

// 	// Компилируем регулярное выражение
// 	re, err := regexp.Compile(pattern)
// 	if err != nil {
// 		return false, err
// 	}

// 	// Проверяем соответствие
// 	return re.MatchString(text), nil
// }

// // MatchesKeywords проверяет, содержит ли сообщение все ключевые слова
// func (s *Service) MatchesKeywords(text string, keywords []string) bool {
// 	if len(keywords) == 0 {
// 		return true
// 	}

// 	// Приводим текст к нижнему регистру для регистронезависимого поиска
// 	lowerText := strings.ToLower(text)

// 	// Проверяем наличие всех ключевых слов
// 	for _, keyword := range keywords {
// 		if !strings.Contains(lowerText, strings.ToLower(keyword)) {
// 			return false
// 		}
// 	}

// 	return true
// }

// // MatchesAnyKeyword проверяет, содержит ли сообщение хотя бы одно из ключевых слов
// func (s *Service) MatchesAnyKeyword(text string, keywords []string) bool {
// 	if len(keywords) == 0 {
// 		return true
// 	}

// 	// Приводим текст к нижнему регистру для регистронезависимого поиска
// 	lowerText := strings.ToLower(text)

// 	// Проверяем наличие хотя бы одного ключевого слова
// 	for _, keyword := range keywords {
// 		if strings.Contains(lowerText, strings.ToLower(keyword)) {
// 			return true
// 		}
// 	}

// 	return false
// }

// // MatchesHashtags проверяет, содержит ли сообщение указанные хэштеги
// func (s *Service) MatchesHashtags(text string, hashtags []string) bool {
// 	if len(hashtags) == 0 {
// 		return true
// 	}

// 	// Компилируем регулярное выражение для поиска хэштегов
// 	re := regexp.MustCompile(`#\w+`)
// 	foundHashtags := re.FindAllString(text, -1)

// 	// Проверяем наличие требуемых хэштегов
// 	for _, requestedTag := range hashtags {
// 		// Добавляем # к тегу, если его нет
// 		if !strings.HasPrefix(requestedTag, "#") {
// 			requestedTag = "#" + requestedTag
// 		}

// 		found := false
// 		for _, foundTag := range foundHashtags {
// 			if strings.EqualFold(foundTag, requestedTag) {
// 				found = true
// 				break
// 			}
// 		}

// 		if !found {
// 			return false
// 		}
// 	}

// 	return true
// }

// // FilterByContentType фильтрует сообщения по типу содержимого
// func (s *Service) FilterByContentType(text string, allowedTypes []string) bool {
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
func (s *Service) ShouldForward(text string, rule *entity.ForwardRule) (bool, error) {
	// Проверка по исключающему регулярному выражению
	if rule.ExcludeRegexp != nil {
		if rule.ExcludeRegexp.MatchString(text) {
			return false, nil
		}
	}

	// Проверка по включающему регулярному выражению
	if rule.IncludeRegexp != nil {
		if !rule.IncludeRegexp.MatchString(text) {
			return false, nil
		}
	}

	// Проверка по подстрокам
	if len(rule.IncludeSubmatch) > 0 {
		matchesAny := false
		for _, submatch := range rule.IncludeSubmatch {
			if submatch.CompiledRegexp != nil {
				matches := submatch.CompiledRegexp.FindStringSubmatch(text) // TODO: зачем внутри цикла?
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
