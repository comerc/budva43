package engine

import (
	"context"
	"fmt"
	"regexp"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/entity"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	GetClientDone() <-chan any
	// tdlibClient methods
	GetListener() *client.Listener
}

//go:generate mockery --name=updateNewMessageHandler --exported
type updateNewMessageHandler interface {
	Run(ctx context.Context, update *client.UpdateNewMessage)
}

//go:generate mockery --name=updateMessageEditedHandler --exported
type updateMessageEditedHandler interface {
	Run(update *client.UpdateMessageEdited)
}

//go:generate mockery --name=updateDeleteMessagesHandler --exported
type updateDeleteMessagesHandler interface {
	Run(update *client.UpdateDeleteMessages)
}

//go:generate mockery --name=updateMessageSendHandler --exported
type updateMessageSendHandler interface {
	Run(update *client.UpdateMessageSendSucceeded)
}

// Service предоставляет функциональность движка пересылки сообщений
type Service struct {
	log *log.Logger
	ctx context.Context
	//
	telegramRepo                telegramRepo
	updateNewMessageHandler     updateNewMessageHandler
	updateMessageEditedHandler  updateMessageEditedHandler
	updateDeleteMessagesHandler updateDeleteMessagesHandler
	updateMessageSendHandler    updateMessageSendHandler
}

// New создает новый экземпляр сервиса engine
func New(
	telegramRepo telegramRepo,
	updateNewMessageHandler updateNewMessageHandler,
	updateMessageEditedHandler updateMessageEditedHandler,
	updateDeleteMessagesHandler updateDeleteMessagesHandler,
	updateMessageSendHandler updateMessageSendHandler,
) *Service {
	return &Service{
		log: log.NewLogger("service.engine"),
		//
		telegramRepo:                telegramRepo,
		updateNewMessageHandler:     updateNewMessageHandler,
		updateMessageEditedHandler:  updateMessageEditedHandler,
		updateDeleteMessagesHandler: updateDeleteMessagesHandler,
		updateMessageSendHandler:    updateMessageSendHandler,
	}
}

// Start запускает обработчик обновлений от Telegram
func (s *Service) Start(ctx context.Context) error {

	s.ctx = ctx

	if err := s.validateConfig(); err != nil {
		return err
	}

	s.transformConfig()

	s.enrichConfig()

	go s.run()

	return nil
}

// Close останавливает сервис
func (s *Service) Close() error {
	return nil
}

// validateConfig проверяет корректность конфигурации
func (s *Service) validateConfig() error {
	if len(config.Engine.Sources) == 0 {
		return log.NewError("отсутствуют настройки",
			"path", "config.Engine.Sources",
		)
	}
	for srcChatId, src := range config.Engine.Sources {
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
	if len(config.Engine.Destinations) == 0 {
		return log.NewError("отсутствуют настройки",
			"path", "config.Engine.Destinations",
		)
	}
	for dstChatId, dsc := range config.Engine.Destinations {
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
	if len(config.Engine.ForwardRules) == 0 {
		return log.NewError("отсутствуют настройки",
			"path", "config.Engine.ForwardRules",
		)
	}
	re := regexp.MustCompile("[:,]") // TODO: зачем нужна эта проверка? (предположительно для badger)
	for forwardRuleId, forwardRule := range config.Engine.ForwardRules {
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

// transformConfig преобразует конфигурацию в отрицательные идентификаторы
func (s *Service) transformConfig() {
	for srcChatId, src := range config.Engine.Sources {
		config.Engine.Sources[-srcChatId] = src
		delete(config.Engine.Sources, srcChatId)
	}
	for _, src := range config.Engine.Sources {
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
	for dstChatId, dsc := range config.Engine.Destinations {
		config.Engine.Destinations[-dstChatId] = dsc
		delete(config.Engine.Destinations, dstChatId)
	}
	for _, forwardRule := range config.Engine.ForwardRules {
		forwardRule.From = -forwardRule.From
		for i, dstChatId := range forwardRule.To {
			forwardRule.To[i] = -dstChatId
		}
		forwardRule.Check = -forwardRule.Check
		forwardRule.Other = -forwardRule.Other
	}
}

// enrichConfig обогащает конфигурацию
func (s *Service) enrichConfig() {
	config.Engine.UniqueSources = make(map[entity.ChatId]struct{})
	tmpOrderedForwardRules := make([]entity.ForwardRuleId, 0)
	for key, destination := range config.Engine.Destinations {
		destination.ChatId = key
	}
	for key, source := range config.Engine.Sources {
		source.ChatId = key
	}
	for key, forwardRule := range config.Engine.ForwardRules {
		forwardRule.Id = key
		if _, ok := config.Engine.Sources[forwardRule.From]; !ok {
			config.Engine.Sources[forwardRule.From] = &entity.Source{
				ChatId: forwardRule.From,
			}
		}
		config.Engine.UniqueSources[forwardRule.From] = struct{}{}
		tmpOrderedForwardRules = append(tmpOrderedForwardRules, forwardRule.Id)
	}
	config.Engine.OrderedForwardRules = util.Distinct(tmpOrderedForwardRules)
}

// run запускает обработчик обновлений от Telegram
func (s *Service) run() {
	// Ждём авторизации клиента и получаем канал обновлений от Telegram
	select {
	case <-s.ctx.Done():
		return
	case <-s.telegramRepo.GetClientDone():
		listener := s.telegramRepo.GetListener()
		defer listener.Close()
		s.handleUpdates(listener)
	}
}

// handleUpdates обрабатывает обновления от Telegram
func (s *Service) handleUpdates(listener *client.Listener) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case update, ok := <-listener.Updates:
			if !ok {
				return
			}

			if update.GetClass() != client.ClassUpdate {
				continue
			}

			switch updateByType := update.(type) {
			case *client.UpdateNewMessage:
				s.updateNewMessageHandler.Run(s.ctx, updateByType)
			case *client.UpdateMessageEdited:
				s.updateMessageEditedHandler.Run(updateByType)
			case *client.UpdateDeleteMessages:
				s.updateDeleteMessagesHandler.Run(updateByType)
			case *client.UpdateMessageSendSucceeded:
				s.updateMessageSendHandler.Run(updateByType)
			}
		}
	}
}
