package engine_config

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/entity"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

var (
	engineViper *viper.Viper
)

func initEngineViper(projectRoot string) {
	engineViper = viper.New()
	path := filepath.Join(projectRoot, "config", config.General.EngineConfigFile)
	engineViper.SetConfigFile(path)
}

type initializeDestinations = func([]entity.ChatId)

// Reload перезагружает конфигурацию config.Engine из engine.yml
func Reload(initializeDestinations initializeDestinations) error {
	newEngineConfig, err := load()
	if err != nil {
		var emptyConfigData *ErrEmptyConfigData
		if !errors.As(err, &emptyConfigData) {
			return log.WrapError(err)
		}
	}

	var destinations []entity.ChatId
	for dstChatId := range newEngineConfig.UniqueDestinations {
		destinations = append(destinations, dstChatId)
	}
	initializeDestinations(destinations)

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
func load() (*entity.EngineConfig, error) {
	if err := engineViper.ReadInConfig(); err != nil {
		return nil, log.WrapError(err)
	}

	engineConfig := &entity.EngineConfig{}
	if err := engineViper.Unmarshal(engineConfig, util.GetConfigOptions()); err != nil {
		return nil, log.WrapError(err)
	}

	Initialize(engineConfig)

	if err := validate(engineConfig); err != nil {
		return nil, log.WrapError(err)
	}

	transform(engineConfig)

	enrich(engineConfig)

	if err := check(engineConfig); err != nil {
		return engineConfig, log.WrapError(err)
	}

	return engineConfig, nil
}

// Initialize инициализирует конфигурацию
func Initialize(engineConfig *entity.EngineConfig) {
	if engineConfig.Sources == nil {
		engineConfig.Sources = make(map[entity.ChatId]*entity.Source)
	}
	if engineConfig.Destinations == nil {
		engineConfig.Destinations = make(map[entity.ChatId]*entity.Destination)
	}
	if engineConfig.ForwardRules == nil {
		engineConfig.ForwardRules = make(map[entity.ForwardRuleId]*entity.ForwardRule)
	}
	engineConfig.UniqueSources = make(map[entity.ChatId]struct{})
	engineConfig.UniqueDestinations = make(map[entity.ChatId]struct{})
}

// validate проверяет корректность конфигурации
func validate(engineConfig *entity.EngineConfig) error {
	for srcChatId, src := range engineConfig.Sources {
		// viper читает цифровые ключи без минуса
		// if srcChatId < 0 {
		// 	return log.NewError("идентификатор не может быть отрицательным",
		// 		"path", "config.Engine.Sources",
		// 		"value", srcChatId)
		// }
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
	}

	for dstChatId, dsc := range engineConfig.Destinations {
		// viper читает цифровые ключи без минуса
		// if dstChatId < 0 {
		// 	return log.NewError("идентификатор не может быть отрицательным",
		// 		"path", "config.Engine.Destinations",
		// 		"value", dstChatId)
		// }
		for i, replaceFragment := range dsc.ReplaceFragments {
			if util.RuneCountForUTF16(replaceFragment.From) != util.RuneCountForUTF16(replaceFragment.To) {
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
func transform(engineConfig *entity.EngineConfig) {
	// Сначала собираем все ключи, чтобы избежать модификации карты во время итерации
	sourceKeys := make([]entity.ChatId, 0, len(engineConfig.Sources))
	for srcChatId := range engineConfig.Sources {
		sourceKeys = append(sourceKeys, srcChatId)
	}

	for _, srcChatId := range sourceKeys {
		src := engineConfig.Sources[srcChatId]
		engineConfig.Sources[-srcChatId] = src
		delete(engineConfig.Sources, srcChatId)
	}

	for _, src := range engineConfig.Sources {
		if src.Sign != nil {
			a := []entity.ChatId{}
			for _, targetChatId := range src.Sign.For {
				a = append(a, -targetChatId)
			}
			src.Sign.For = a
		}
		if src.Link != nil {
			a := []entity.ChatId{}
			for _, targetChatId := range src.Link.For {
				a = append(a, -targetChatId)
			}
			src.Link.For = a
		}
	}

	// Сначала собираем все ключи, чтобы избежать модификации карты во время итерации
	destinationKeys := make([]entity.ChatId, 0, len(engineConfig.Destinations))
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
func enrich(engineConfig *entity.EngineConfig) {
	tmpOrderedForwardRules := make([]entity.ForwardRuleId, 0)

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
			engineConfig.Sources[srcChatId] = &entity.Source{
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

type ErrEmptyConfigData struct {
	log.CustomError
}

// check проверяет, что конфигурация не пуста
func check(engineConfig *entity.EngineConfig) error {
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
		return &ErrEmptyConfigData{
			CustomError: *log.NewError("отсутствуют данные", args...),
		}
	}

	return nil
}
