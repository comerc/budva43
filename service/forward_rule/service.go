package forward_rule

import (
	"regexp"

	"github.com/comerc/budva43/entity"
)

// ForwardRuleService предоставляет методы для работы с правилами пересылки
type ForwardRuleService struct {
	// Здесь могут быть зависимости, например, репозитории
}

// NewForwardRuleService создает новый экземпляр сервиса для работы с правилами пересылки
func NewForwardRuleService() *ForwardRuleService {
	return &ForwardRuleService{}
}

// CompileRegexps компилирует все регулярные выражения в правиле
func (s *ForwardRuleService) CompileRegexps(rule *entity.ForwardRule) error {
	var err error

	// Компилируем регулярное выражение для исключения
	if rule.Exclude != "" {
		rule.ExcludeRegexp, err = regexp.Compile(rule.Exclude)
		if err != nil {
			return err
		}
	}

	// Компилируем регулярное выражение для включения
	if rule.Include != "" {
		rule.IncludeRegexp, err = regexp.Compile(rule.Include)
		if err != nil {
			return err
		}
	}

	// Компилируем регулярные выражения для правил подстрок
	for i := range rule.IncludeSubmatch {
		rule.IncludeSubmatch[i].CompiledRegexp, err = regexp.Compile(rule.IncludeSubmatch[i].Regexp)
		if err != nil {
			return err
		}
	}

	return nil
}

// ShouldForward проверяет, должно ли сообщение быть переслано согласно правилу
func (s *ForwardRuleService) ShouldForward(rule *entity.ForwardRule, text string) bool {
	// Если правило неактивно, не пересылаем сообщение
	if rule.Status != entity.RuleStatusActive {
		return false
	}

	// Если есть регулярное выражение для исключения и оно совпадает с текстом,
	// не пересылаем сообщение
	if rule.ExcludeRegexp != nil && rule.ExcludeRegexp.MatchString(text) {
		return false
	}

	// Если есть регулярное выражение для включения, проверяем его
	if rule.IncludeRegexp != nil {
		// Если не совпадает, не пересылаем сообщение
		if !rule.IncludeRegexp.MatchString(text) {
			return false
		}
	}

	// Проверяем правила для подстрок
	for _, submatchRule := range rule.IncludeSubmatch {
		if submatchRule.CompiledRegexp != nil {
			matches := submatchRule.CompiledRegexp.FindStringSubmatch(text)

			// Если нет совпадений или группа за пределами количества совпадений,
			// переходим к следующему правилу
			if len(matches) <= submatchRule.Group {
				continue
			}

			// Проверяем, есть ли подстрока в списке допустимых значений
			match := false
			for _, allowedMatch := range submatchRule.Match {
				if matches[submatchRule.Group] == allowedMatch {
					match = true
					break
				}
			}

			// Если нет совпадения в списке, не пересылаем сообщение
			if !match {
				return false
			}
		}
	}

	// Если все проверки пройдены, сообщение может быть переслано
	return true
}
