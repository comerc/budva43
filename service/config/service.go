package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

//go:generate mockery --name=configLoader --exported
type configLoader interface {
	LoadConfig(path string) ([]byte, error)
}

// DefaultConfigLoader реализация загрузчика конфигурации по умолчанию
type DefaultConfigLoader struct{}

// LoadConfig загружает конфигурацию из файла
func (l *DefaultConfigLoader) LoadConfig(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

//go:generate mockery --name=configStorage --exported
type configStorage interface {
	SaveConfig(name string, data []byte) error
}

// ConfigService предоставляет методы для работы с конфигурацией
type ConfigService struct {
	loader      configLoader
	storage     configStorage
	configPath  string
	config      map[string]interface{}
	configLock  sync.RWMutex
	lastModTime time.Time
}

// NewConfigService создает новый экземпляр сервиса для работы с конфигурацией
func NewConfigService(configPath string, loader configLoader, storage configStorage) *ConfigService {
	if loader == nil {
		loader = &DefaultConfigLoader{}
	}

	return &ConfigService{
		loader:     loader,
		storage:    storage,
		configPath: configPath,
		config:     make(map[string]interface{}),
	}
}

// LoadConfig загружает конфигурацию из файла
func (s *ConfigService) LoadConfig() error {
	// Проверяем существование файла
	info, err := os.Stat(s.configPath)
	if err != nil {
		return err
	}

	// Если файл не изменился с момента последней загрузки, ничего не делаем
	if !info.ModTime().After(s.lastModTime) && len(s.config) > 0 {
		return nil
	}

	// Загружаем конфигурацию
	data, err := s.loader.LoadConfig(s.configPath)
	if err != nil {
		return err
	}

	// Разбираем JSON
	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	// Обновляем конфигурацию
	s.configLock.Lock()
	s.config = config
	s.lastModTime = info.ModTime()
	s.configLock.Unlock()

	// Если есть хранилище, сохраняем в него
	if s.storage != nil {
		err = s.storage.SaveConfig("current", data)
		if err != nil {
			// Нелитричная ошибка, можем продолжить
			// Логировать ошибку
		}
	}

	return nil
}

// GetConfigValue получает значение из конфигурации по ключу
func (s *ConfigService) GetConfigValue(key string) (interface{}, error) {
	s.configLock.RLock()
	defer s.configLock.RUnlock()

	if len(s.config) == 0 {
		return nil, errors.New("config is not loaded")
	}

	value, ok := s.config[key]
	if !ok {
		return nil, errors.New("key not found in config")
	}

	return value, nil
}

// SetConfigValue устанавливает значение в конфигурации по ключу
func (s *ConfigService) SetConfigValue(key string, value interface{}) {
	s.configLock.Lock()
	defer s.configLock.Unlock()

	s.config[key] = value
}

// SaveConfig сохраняет текущую конфигурацию в файл
func (s *ConfigService) SaveConfig() error {
	s.configLock.RLock()
	defer s.configLock.RUnlock()

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(s.configPath, data, 0644)
	if err != nil {
		return err
	}

	// Если есть хранилище, сохраняем в него
	if s.storage != nil {
		err = s.storage.SaveConfig("current", data)
		if err != nil {
			// Нелитричная ошибка, можем продолжить
			// Логировать ошибку
		}
	}

	return nil
}

// IsConfigLoaded проверяет, загружена ли конфигурация
func (s *ConfigService) IsConfigLoaded() bool {
	s.configLock.RLock()
	defer s.configLock.RUnlock()
	return len(s.config) > 0
}

// WatchConfigChanges запускает мониторинг изменений конфигурации
func (s *ConfigService) WatchConfigChanges(interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := s.LoadConfig()
			if err != nil {
				// Логировать ошибку
			}
		case <-stopCh:
			return
		}
	}
}
