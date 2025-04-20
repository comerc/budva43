package auto_answer

import (
	"log/slog"
	"regexp"
	"strings"
	"sync"

	"github.com/zelenin/go-tdlib/client"
)

//go:generate mockery --name=messageProcessor --exported
type messageProcessor interface {
	GetText(message *client.Message) string
	GetCaption(message *client.Message) string
}

// MessageMatcher интерфейс для сопоставления сообщений
type MessageMatcher interface {
	Match(message string) bool
}

// RegexpMatcher реализация сопоставления по регулярному выражению
type RegexpMatcher struct {
	pattern *regexp.Regexp
}

// NewRegexpMatcher создает новый сопоставитель по регулярному выражению
func NewRegexpMatcher(pattern string) (*RegexpMatcher, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &RegexpMatcher{pattern: re}, nil
}

// Match проверяет соответствие сообщения шаблону
func (m *RegexpMatcher) Match(message string) bool {
	return m.pattern.MatchString(message)
}

// KeywordMatcher реализация сопоставления по ключевым словам
type KeywordMatcher struct {
	keywords []string
}

// NewKeywordMatcher создает новый сопоставитель по ключевым словам
func NewKeywordMatcher(keywords []string) *KeywordMatcher {
	return &KeywordMatcher{keywords: keywords}
}

// Match проверяет наличие ключевых слов в сообщении
func (m *KeywordMatcher) Match(message string) bool {
	lowerMessage := strings.ToLower(message)
	for _, keyword := range m.keywords {
		if strings.Contains(lowerMessage, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// Rule правило для автоматического ответа
type Rule struct {
	Name         string
	Matcher      MessageMatcher
	Response     string
	Priority     int
	OnlyPrivate  bool
	OnlyInGroups bool
	Enabled      bool
}

// Service предоставляет методы для автоматических ответов
type Service struct {
	log *slog.Logger
	//
	messageProcessor messageProcessor
	rules            []*Rule
	rulesByName      map[string]*Rule
	mutex            sync.RWMutex
}

// New создает новый экземпляр сервиса для автоматических ответов
func New(messageProcessor messageProcessor) *Service {
	return &Service{
		log: slog.With("module", "service.auto_answer"),
		//
		messageProcessor: messageProcessor,
		rules:            make([]*Rule, 0),
		rulesByName:      make(map[string]*Rule),
	}
}

// AddRule добавляет правило автоответа
func (s *Service) AddRule(rule *Rule) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Проверяем, существует ли правило с таким именем
	if existingRule, exists := s.rulesByName[rule.Name]; exists {
		// Обновляем существующее правило
		existingRule.Matcher = rule.Matcher
		existingRule.Response = rule.Response
		existingRule.Priority = rule.Priority
		existingRule.OnlyPrivate = rule.OnlyPrivate
		existingRule.OnlyInGroups = rule.OnlyInGroups
		existingRule.Enabled = rule.Enabled
	} else {
		// Добавляем новое правило
		s.rules = append(s.rules, rule)
		s.rulesByName[rule.Name] = rule
	}

	// Сортируем правила по приоритету (высший приоритет в начале)
	s.sortRules()
}

// RemoveRule удаляет правило автоответа
func (s *Service) RemoveRule(name string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.rulesByName[name]; !exists {
		return false
	}

	// Удаляем правило из массива
	for i, rule := range s.rules {
		if rule.Name == name {
			s.rules = append(s.rules[:i], s.rules[i+1:]...)
			break
		}
	}

	// Удаляем правило из карты
	delete(s.rulesByName, name)

	return true
}

// EnableRule включает правило автоответа
func (s *Service) EnableRule(name string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	rule, exists := s.rulesByName[name]
	if !exists {
		return false
	}

	rule.Enabled = true
	return true
}

// DisableRule выключает правило автоответа
func (s *Service) DisableRule(name string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	rule, exists := s.rulesByName[name]
	if !exists {
		return false
	}

	rule.Enabled = false
	return true
}

// ProcessMessage обрабатывает сообщение и возвращает автоответ, если есть подходящее правило
func (s *Service) ProcessMessage(message *client.Message, isPrivate bool) (string, bool) {
	if message == nil {
		return "", false
	}

	// Получаем текст сообщения
	text := s.messageProcessor.GetText(message)
	if text == "" {
		text = s.messageProcessor.GetCaption(message)
	}
	if text == "" {
		return "", false
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Проверяем каждое правило
	for _, rule := range s.rules {
		// Пропускаем отключенные правила
		if !rule.Enabled {
			continue
		}

		// Проверяем ограничения по типу чата
		if (rule.OnlyPrivate && !isPrivate) || (rule.OnlyInGroups && isPrivate) {
			continue
		}

		// Проверяем соответствие сообщения шаблону
		if rule.Matcher.Match(text) {
			return rule.Response, true
		}
	}

	return "", false
}

// GetAllRules возвращает все правила автоответов
func (s *Service) GetAllRules() []*Rule {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Создаем копию, чтобы избежать параллельного доступа
	result := make([]*Rule, len(s.rules))
	copy(result, s.rules)

	return result
}

// GetRule возвращает правило по имени
func (s *Service) GetRule(name string) (*Rule, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	rule, exists := s.rulesByName[name]
	return rule, exists
}

// sortRules сортирует правила по приоритету
func (s *Service) sortRules() {
	// Сортировка пузырьком для простоты (для небольшого количества правил)
	for i := 0; i < len(s.rules)-1; i++ {
		for j := 0; j < len(s.rules)-i-1; j++ {
			if s.rules[j].Priority < s.rules[j+1].Priority {
				s.rules[j], s.rules[j+1] = s.rules[j+1], s.rules[j]
			}
		}
	}
}
