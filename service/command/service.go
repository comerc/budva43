package command

import (
	"errors"
	"strings"
	"sync"
)

//go:generate mockery --name=commandHandler --exported
type commandHandler interface {
	HandleCommand(command string, args []string) (string, error)
}

// Service предоставляет методы для обработки команд
type Service struct {
	handlers     map[string]commandHandler
	aliases      map[string]string
	helpMessages map[string]string
	mutex        sync.RWMutex
}

// New создает новый экземпляр сервиса для обработки команд
func New() *Service {
	return &Service{
		handlers:     make(map[string]commandHandler),
		aliases:      make(map[string]string),
		helpMessages: make(map[string]string),
	}
}

// RegisterHandler регистрирует обработчик для команды
func (s *Service) RegisterHandler(command string, handler commandHandler, help string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.handlers[command] = handler
	if help != "" {
		s.helpMessages[command] = help
	}
}

// RegisterAlias регистрирует псевдоним для команды
func (s *Service) RegisterAlias(alias, command string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.handlers[command]; !exists {
		return errors.New("command not registered")
	}

	s.aliases[alias] = command
	return nil
}

// ExecuteCommand выполняет команду
func (s *Service) ExecuteCommand(input string) (string, error) {
	// Разбираем ввод на команду и аргументы
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", errors.New("empty command")
	}

	command := strings.ToLower(parts[0])
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Проверяем, является ли это псевдонимом
	if aliasCommand, isAlias := s.aliases[command]; isAlias {
		command = aliasCommand
	}

	// Ищем обработчик
	handler, exists := s.handlers[command]
	if !exists {
		return "", errors.New("unknown command: " + command)
	}

	// Выполняем команду
	return handler.HandleCommand(command, args)
}

// GetCommandList возвращает список доступных команд
func (s *Service) GetCommandList() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	commands := make([]string, 0, len(s.handlers))
	for command := range s.handlers {
		commands = append(commands, command)
	}

	return commands
}

// GetCommandHelp возвращает справку по команде
func (s *Service) GetCommandHelp(command string) string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Проверяем, является ли это псевдонимом
	if aliasCommand, isAlias := s.aliases[command]; isAlias {
		command = aliasCommand
	}

	help, exists := s.helpMessages[command]
	if !exists {
		return "No help available for this command"
	}

	return help
}

// GetAllHelpMessages возвращает справку по всем командам
func (s *Service) GetAllHelpMessages() map[string]string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Создаем копию, чтобы избежать параллельного доступа
	result := make(map[string]string, len(s.helpMessages))
	for command, help := range s.helpMessages {
		result[command] = help
	}

	return result
}

// HasCommand проверяет наличие команды
func (s *Service) HasCommand(command string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Проверяем, является ли это псевдонимом
	if aliasCommand, isAlias := s.aliases[command]; isAlias {
		command = aliasCommand
	}

	_, exists := s.handlers[command]
	return exists
}

// UnregisterHandler удаляет обработчик команды
func (s *Service) UnregisterHandler(command string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.handlers, command)
	delete(s.helpMessages, command)

	// Удаляем все псевдонимы, ссылающиеся на эту команду
	for alias, cmd := range s.aliases {
		if cmd == command {
			delete(s.aliases, alias)
		}
	}
}

// ParseCommand разбирает строку команды на команду и аргументы
func (s *Service) ParseCommand(input string) (string, []string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", nil
	}

	command := strings.ToLower(parts[0])
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	return command, args
}
