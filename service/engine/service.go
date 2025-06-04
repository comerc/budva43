package engine

import (
	"context"
	"regexp"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/entity"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

type telegramRepo interface {
	GetClient() *client.Client
	GetClientDone() <-chan any
}

type updateNewMessageHandler interface {
	Run(ctx context.Context, update *client.UpdateNewMessage)
}

type updateMessageEditedHandler interface {
	Run(update *client.UpdateMessageEdited)
}

type updateDeleteMessagesHandler interface {
	Run(update *client.UpdateDeleteMessages)
}

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

	return nil

	// Проверяем конфигурацию
	if err := s.validateConfig(); err != nil {
		return err
	}

	// Обогащаем конфигурацию
	if err := s.enrichConfig(); err != nil {
		return err
	}

	go s.run()

	return nil
}

// Close останавливает сервис
func (s *Service) Close() error {
	return nil
}

// validateConfig проверяет корректность конфигурации
func (s *Service) validateConfig() error {
	for _, dsc := range config.Engine.Destinations {
		for _, replaceFragment := range dsc.ReplaceFragments {
			if util.RuneCountForUTF16(replaceFragment.From) != util.RuneCountForUTF16(replaceFragment.To) {
				return log.NewError("длина исходного и заменяемого текста должна быть одинаковой",
					"from", replaceFragment.From,
					"to", replaceFragment.To,
				)
			}
		}
	}

	re := regexp.MustCompile("[:,]") // TODO: зачем нужна эта проверка? (предположительно для badger)
	for forwardRuleId, forwardRule := range config.Engine.ForwardRules {
		if re.FindString(forwardRuleId) != "" {
			return log.NewError("нельзя использовать [:,] в идентификаторе правила",
				"forwardRuleId", forwardRuleId,
			)
		}
		for _, dstChatId := range forwardRule.To {
			if forwardRule.From == dstChatId {
				return log.NewError("идентификатор получателя не может совпадать с идентификатором источника",
					"dstChatId", dstChatId)
			}
		}
	}

	return nil
}

// enrichConfig обогащает конфигурацию
func (s *Service) enrichConfig() error {
	if len(config.Engine.Destinations) == 0 {
		return log.NewError("отсутствуют настройки получателей")
	}
	if len(config.Engine.Sources) == 0 {
		return log.NewError("отсутствуют настройки источников")
	}
	if len(config.Engine.ForwardRules) == 0 {
		return log.NewError("отсутствуют настройки пересылки")
	}

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
	return nil
}

// run запускает обработчик обновлений от Telegram
func (s *Service) run() {
	// Ждём авторизации клиента и получаем канал обновлений от Telegram
	select {
	case <-s.ctx.Done():
		return
	case <-s.telegramRepo.GetClientDone():
		listener := s.telegramRepo.GetClient().GetListener()
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
