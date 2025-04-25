package filter

import (
	"log/slog"

	"github.com/comerc/budva43/entity"
)

// TODO: в старой реализации checkFilters имел более сложную логику с проверкой Include/Exclude и регулярными выражениями

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
