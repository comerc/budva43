package engine_config

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/entity"
	"github.com/comerc/budva43/app/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/comerc/budva43/app/util"
)

var (
	engineViper *viper.Viper
)

func initEngineViper(projectRoot string) {
	engineViper = viper.New()
	engineViper.SetConfigName("engine")
	engineViper.SetConfigType("yml")
	engineViper.AddConfigPath(filepath.Join(projectRoot, "config"))
}

// Reload перезагружает конфигурацию config.Engine из engine.yml
func Reload() error {
	newEngineConfig, err := load()
	if err != nil {
		return err // log.WrapError(err) - уже есть внутри load()
	}

	// Атомарно заменяем глобальную конфигурацию
	config.Engine = newEngineConfig

	return nil
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

	config := &entity.EngineConfig{}
	if err := engineViper.Unmarshal(config, util.GetConfigOptions()); err != nil {
		return nil, log.WrapError(err)
	}

	if err := validate(config); err != nil {
		return nil, log.WrapError(err)
	}

	transform(config)

	enrich(config)

	return config, nil
}

// validate проверяет корректность конфигурации движка
func validate(config *entity.EngineConfig) error {
	if len(config.Sources) == 0 {
		return log.NewError("отсутствуют настройки",
			"path", "config.Engine.Sources",
		)
	}

	for srcChatId, src := range config.Sources {
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

	if len(config.Destinations) == 0 {
		return log.NewError("отсутствуют настройки", "path", "config.Engine.Destinations")
	}

	for dstChatId, dsc := range config.Destinations {
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

	if len(config.ForwardRules) == 0 {
		return log.NewError("отсутствуют настройки", "path", "config.Engine.ForwardRules")
	}

	re := regexp.MustCompile("[:,]") // TODO: зачем нужна эта проверка? (предположительно для badger)
	for forwardRuleId, forwardRule := range config.ForwardRules {
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
func transform(config *entity.EngineConfig) {
	// Сначала собираем все ключи, чтобы избежать модификации карты во время итерации
	sourceKeys := make([]entity.ChatId, 0, len(config.Sources))
	for srcChatId := range config.Sources {
		sourceKeys = append(sourceKeys, srcChatId)
	}

	for _, srcChatId := range sourceKeys {
		src := config.Sources[srcChatId]
		config.Sources[-srcChatId] = src
		delete(config.Sources, srcChatId)
	}

	for _, src := range config.Sources {
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
	destinationKeys := make([]entity.ChatId, 0, len(config.Destinations))
	for dstChatId := range config.Destinations {
		destinationKeys = append(destinationKeys, dstChatId)
	}

	for _, dstChatId := range destinationKeys {
		dsc := config.Destinations[dstChatId]
		config.Destinations[-dstChatId] = dsc
		delete(config.Destinations, dstChatId)
	}

	for _, forwardRule := range config.ForwardRules {
		forwardRule.From = -forwardRule.From
		for i, dstChatId := range forwardRule.To {
			forwardRule.To[i] = -dstChatId
		}
		forwardRule.Check = -forwardRule.Check
		forwardRule.Other = -forwardRule.Other
	}
}

// enrich обогащает конфигурацию
func enrich(config *entity.EngineConfig) {
	config.UniqueSources = make(map[entity.ChatId]struct{})
	tmpOrderedForwardRules := make([]entity.ForwardRuleId, 0)

	for key, destination := range config.Destinations {
		destination.ChatId = key
	}

	for key, source := range config.Sources {
		source.ChatId = key
	}

	for key, forwardRule := range config.ForwardRules {
		forwardRule.Id = key
		if _, ok := config.Sources[forwardRule.From]; !ok {
			config.Sources[forwardRule.From] = &entity.Source{
				ChatId: forwardRule.From,
			}
		}
		config.UniqueSources[forwardRule.From] = struct{}{}
		tmpOrderedForwardRules = append(tmpOrderedForwardRules, forwardRule.Id)
	}

	config.OrderedForwardRules = util.Distinct(tmpOrderedForwardRules)
}
