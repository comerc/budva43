package engine_config

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/domain"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

var (
	engineViper *viper.Viper
)

func initEngineViper(projectRoot string) {
	engineViper = viper.New()
	path := filepath.Join(projectRoot, ".config", config.General.EngineConfigFile)
	engineViper.SetConfigFile(path)
}

type initDestinations = func([]domain.ChatId)

// Reload перезагружает конфигурацию config.Engine из engine.yml
func Reload(initDestinations initDestinations) error {
	newEngineConfig, err := load()
	if err != nil {
		if !errors.Is(err, ErrEmptyConfigData) {
			return err
		}
	}

	var destinations []domain.ChatId
	for dstChatId := range newEngineConfig.UniqueDestinations {
		destinations = append(destinations, dstChatId)
	}
	initDestinations(destinations)

	// Атомарно заменяем глобальную конфигурацию
	config.Engine = newEngineConfig

	return err // нужно вернуть ErrEmptyConfigData
}

// Watch настраивает отслеживание изменений engine.yml
func Watch(reloadCallback func()) {
	engineViper.OnConfigChange(func(e fsnotify.Event) {
		reloadCallback()
	})
	engineViper.WatchConfig()
}

// load загружает конфигурацию из engine.yml
func load() (*domain.EngineConfig, error) {
	if err := engineViper.ReadInConfig(); err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}

	engineConfig := &domain.EngineConfig{}
	if err := engineViper.Unmarshal(engineConfig, util.GetConfigOptions()); err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}

	Initialize(engineConfig)

	if err := validate(engineConfig); err != nil {
		return nil, err
	}

	transform(engineConfig)

	enrich(engineConfig)

	if err := check(engineConfig); err != nil {
		return engineConfig, err
	}

	return engineConfig, nil
}

// Initialize инициализирует конфигурацию
func Initialize(engineConfig *domain.EngineConfig) {
	if engineConfig.Sources == nil {
		engineConfig.Sources = make(map[domain.ChatId]*domain.Source)
	}
	if engineConfig.Destinations == nil {
		engineConfig.Destinations = make(map[domain.ChatId]*domain.Destination)
	}
	if engineConfig.ForwardRules == nil {
		engineConfig.ForwardRules = make(map[domain.ForwardRuleId]*domain.ForwardRule)
	}
	engineConfig.UniqueSources = make(map[domain.ChatId]struct{})
	engineConfig.UniqueDestinations = make(map[domain.ChatId]struct{})
}

// validate проверяет корректность конфигурации
func validate(engineConfig *domain.EngineConfig) error {
	for srcChatId, src := range engineConfig.Sources {
		// viper читает цифровые ключи без минуса
		// if srcChatId < 0 {
		// 	return log.NewError("идентификатор не может быть отрицательным",
		// 		"path", "config.Engine.Sources",
		// 		"value", srcChatId)
		// }
		if src.Translate != nil {
			for _, targetChatId := range src.Translate.For {
				if targetChatId < 0 {
					return log.NewError("идентификатор не может быть отрицательным",
						"path", fmt.Sprintf("config.Engine.Sources[%d].Translate.For", srcChatId),
						"value", targetChatId)
				}
			}
		}
		if src.Sign != nil {
			for _, targetChatId := range src.Sign.For {
				if targetChatId < 0 {
					return log.NewError("идентификатор не может быть отрицательным",
						"path", fmt.Sprintf("config.Engine.Sources[%d].Sign.For", srcChatId),
						"value", targetChatId)
				}
			}
		}
		if src.Link != nil {
			for _, targetChatId := range src.Link.For {
				if targetChatId < 0 {
					return log.NewError("идентификатор не может быть отрицательным",
						"path", fmt.Sprintf("config.Engine.Sources[%d].Link.For", srcChatId),
						"value", targetChatId)
				}
			}
		}
		if src.Prev != nil {
			for _, targetChatId := range src.Prev.For {
				if targetChatId < 0 {
					return log.NewError("идентификатор не может быть отрицательным",
						"path", fmt.Sprintf("config.Engine.Sources[%d].Prev.For", srcChatId),
						"value", targetChatId)
				}
			}
		}
		if src.Next != nil {
			for _, targetChatId := range src.Next.For {
				if targetChatId < 0 {
					return log.NewError("идентификатор не может быть отрицательным",
						"path", fmt.Sprintf("config.Engine.Sources[%d].Next.For", srcChatId),
						"value", targetChatId)
				}
			}
		}
	}

	for dstChatId, dsc := range engineConfig.Destinations {
		// viper читает цифровые ключи без минуса
		// if dstChatId < 0 {
		// 	return log.NewError("идентификатор не может быть отрицательным",
		// 		"path", "config.Engine.Destinations",
		// 		"value", dstChatId)
		// }
		for i, replaceFragment := range dsc.ReplaceFragments {
			if len(util.EncodeToUTF16(replaceFragment.From)) != len(util.EncodeToUTF16(replaceFragment.To)) {
				return log.NewError("длина исходного и заменяемого текста должна быть одинаковой",
					"path", fmt.Sprintf("config.Engine.Destinations[%d].ReplaceFragments[%d]", dstChatId, i),
					"from", replaceFragment.From,
					"to", replaceFragment.To,
				)
			}
		}
	}

	re := regexp.MustCompile("[:,]") // TODO: зачем нужна эта проверка? (предположительно для badger)
	for forwardRuleId, forwardRule := range engineConfig.ForwardRules {
		if re.FindString(forwardRuleId) != "" {
			return log.NewError("нельзя использовать [:,] в идентификаторе",
				"path", "config.Engine.ForwardRules",
				"value", forwardRuleId,
			)
		}

		// viper читает именные ключи в PascalCase
		// if cases.Title(language.English).String(forwardRuleId) != forwardRuleId {
		// 	return log.NewError("идентификатор должен быть в PascalCase",
		// 		"path", "config.Engine.ForwardRules",
		// 		"value", forwardRuleId,
		// 	)
		// }

		if forwardRule.From < 0 {
			return log.NewError("идентификатор не может быть отрицательным",
				"path", fmt.Sprintf("config.Engine.ForwardRules[%s].From", forwardRuleId),
				"value", forwardRule.From)
		}

		for i, dstChatId := range forwardRule.To {
			if dstChatId < 0 {
				return log.NewError("идентификатор не может быть отрицательным",
					"path", fmt.Sprintf("config.Engine.ForwardRules[%s].To[%d]", forwardRuleId, i),
					"value", dstChatId)
			}
			if forwardRule.From == dstChatId {
				return log.NewError("идентификатор получателя не может совпадать с идентификатором источника",
					"path", fmt.Sprintf("config.Engine.ForwardRules[%s].To[%d]", forwardRuleId, i),
					"value", dstChatId)
			}
		}

		if forwardRule.Check < 0 {
			return log.NewError("идентификатор не может быть отрицательным",
				"path", fmt.Sprintf("config.Engine.ForwardRules[%s].Check", forwardRuleId),
				"value", forwardRule.Check)
		}

		if forwardRule.Other < 0 {
			return log.NewError("идентификатор не может быть отрицательным",
				"path", fmt.Sprintf("config.Engine.ForwardRules[%s].Other", forwardRuleId),
				"value", forwardRule.Other)
		}
	}

	return nil
}

// transform преобразует конфигурацию в отрицательные идентификаторы
func transform(engineConfig *domain.EngineConfig) {
	// Сначала собираем все ключи, чтобы избежать модификации карты во время итерации
	sourceKeys := make([]domain.ChatId, 0, len(engineConfig.Sources))
	for srcChatId := range engineConfig.Sources {
		sourceKeys = append(sourceKeys, srcChatId)
	}

	for _, srcChatId := range sourceKeys {
		src := engineConfig.Sources[srcChatId]
		engineConfig.Sources[-srcChatId] = src
		delete(engineConfig.Sources, srcChatId)
	}

	for _, src := range engineConfig.Sources {
		if src.Translate != nil {
			a := []domain.ChatId{}
			for _, targetChatId := range src.Translate.For {
				a = append(a, -targetChatId)
			}
			src.Translate.For = a
		}
		if src.Sign != nil {
			a := []domain.ChatId{}
			for _, targetChatId := range src.Sign.For {
				a = append(a, -targetChatId)
			}
			src.Sign.For = a
		}
		if src.Link != nil {
			a := []domain.ChatId{}
			for _, targetChatId := range src.Link.For {
				a = append(a, -targetChatId)
			}
			src.Link.For = a
		}
		if src.Prev != nil {
			a := []domain.ChatId{}
			for _, targetChatId := range src.Prev.For {
				a = append(a, -targetChatId)
			}
			src.Prev.For = a
		}
		if src.Next != nil {
			a := []domain.ChatId{}
			for _, targetChatId := range src.Next.For {
				a = append(a, -targetChatId)
			}
			src.Next.For = a
		}
	}

	// Сначала собираем все ключи, чтобы избежать модификации карты во время итерации
	destinationKeys := make([]domain.ChatId, 0, len(engineConfig.Destinations))
	for dstChatId := range engineConfig.Destinations {
		destinationKeys = append(destinationKeys, dstChatId)
	}

	for _, dstChatId := range destinationKeys {
		dsc := engineConfig.Destinations[dstChatId]
		engineConfig.Destinations[-dstChatId] = dsc
		delete(engineConfig.Destinations, dstChatId)
	}

	for _, forwardRule := range engineConfig.ForwardRules {
		forwardRule.From = -forwardRule.From
		for i, dstChatId := range forwardRule.To {
			forwardRule.To[i] = -dstChatId
		}
		forwardRule.Check = -forwardRule.Check
		forwardRule.Other = -forwardRule.Other
	}
}

// enrich обогащает конфигурацию
func enrich(engineConfig *domain.EngineConfig) {
	tmpOrderedForwardRules := make([]domain.ForwardRuleId, 0)

	for key, destination := range engineConfig.Destinations {
		destination.ChatId = key
	}

	for key, source := range engineConfig.Sources {
		source.ChatId = key
	}

	for key, forwardRule := range engineConfig.ForwardRules {
		srcChatId := forwardRule.From
		forwardRule.Id = key
		if _, ok := engineConfig.Sources[srcChatId]; !ok {
			engineConfig.Sources[srcChatId] = &domain.Source{
				ChatId: srcChatId,
			}
		}
		engineConfig.UniqueSources[srcChatId] = struct{}{}
		for _, dstChatId := range forwardRule.To {
			engineConfig.UniqueDestinations[dstChatId] = struct{}{}
		}
		tmpOrderedForwardRules = append(tmpOrderedForwardRules, forwardRule.Id)
	}

	engineConfig.OrderedForwardRules = util.Distinct(tmpOrderedForwardRules)
}

var ErrEmptyConfigData = errors.New("отсутствуют данные")

// check проверяет, что конфигурация не пуста
func check(engineConfig *domain.EngineConfig) error {
	var args []any

	getKey := util.NewFuncWithIndex("path") // !! частичное применение

	if len(engineConfig.UniqueSources) == 0 {
		args = append(args, getKey(), "config.Engine.UniqueSources")
	}
	if len(engineConfig.UniqueDestinations) == 0 {
		args = append(args, getKey(), "config.Engine.UniqueDestinations")
	}
	if len(engineConfig.OrderedForwardRules) == 0 {
		args = append(args, getKey(), "config.Engine.OrderedForwardRules")
	}

	if len(args) > 0 {
		return log.WrapError(ErrEmptyConfigData, args...)
	}

	return nil
}
