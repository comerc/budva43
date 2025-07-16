package filters_mode

import (
	"regexp"
	"slices"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/domain"
	"github.com/comerc/budva43/app/log"
)

type Service struct {
	log *log.Logger
}

func New() *Service {
	return &Service{
		log: log.NewLogger(),
	}
}

// Map определяет, какой режим фильтрации применим
func (s *Service) Map(formattedText *client.FormattedText, rule *domain.ForwardRule) domain.FiltersMode {
	if formattedText.Text == "" {
		hasInclude := false
		if rule.Include != "" {
			hasInclude = true
		}
		for _, includeSubmatch := range rule.IncludeSubmatch {
			if includeSubmatch.Regexp != "" {
				hasInclude = true
				break
			}
		}
		if hasInclude {
			return domain.FiltersOther
		}
	} else {
		if rule.Exclude != "" {
			re := regexp.MustCompile("(?i)" + rule.Exclude)
			if re.FindString(formattedText.Text) != "" {
				return domain.FiltersCheck
			}
		}
		hasInclude := false
		if rule.Include != "" {
			hasInclude = true
			re := regexp.MustCompile("(?i)" + rule.Include)
			if re.FindString(formattedText.Text) != "" {
				return domain.FiltersOK
			}
		}
		for _, includeSubmatch := range rule.IncludeSubmatch {
			if includeSubmatch.Regexp != "" {
				hasInclude = true
				re := regexp.MustCompile("(?i)" + includeSubmatch.Regexp)
				matches := re.FindAllStringSubmatch(formattedText.Text, -1)
				for _, match := range matches {
					s := match[includeSubmatch.Group]
					if slices.Contains(includeSubmatch.Match, s) {
						return domain.FiltersOK
					}
				}
			}
		}
		if hasInclude {
			return domain.FiltersOther
		}
	}
	return domain.FiltersOK
}
