package loader

import (
	"errors"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/engine_config"
	"github.com/comerc/budva43/app/entity"
	"github.com/comerc/budva43/app/log"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	// tdlibClient methods
	LoadChats(*client.LoadChatsRequest) (*client.Ok, error)
	GetChatHistory(*client.GetChatHistoryRequest) (*client.Messages, error)
}

// Service предоставляет функциональность загрузчика
type Service struct {
	log *log.Logger
	//
	telegramRepo telegramRepo
}

// New создает новый экземпляр сервиса загрузчика
func New(
	telegramRepo telegramRepo,
) *Service {
	return &Service{
		log: log.NewLogger(),
		//
		telegramRepo: telegramRepo,
	}
}

// Run запускает сервис загрузчика
func (s *Service) Run() {
	// Загружаем в первый раз engine.yml
	s.handleConfigReload()

	// Подключаем отслеживание изменений engine.yml
	engine_config.Watch(s.handleConfigReload)
}

// handleConfigReload обрабатывает изменения конфигурации
func (s *Service) handleConfigReload() {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

	err = engine_config.Reload(newFuncInitDestinations(s))

	if errors.Is(err, engine_config.ErrEmptyConfigData) {
		var customError *log.CustomError
		if errors.As(err, &customError) {
			s.log.ErrorOrWarn(nil, err.Error(), customError.Args...)
		}
		err = nil
	}
}

type initDestinations = func([]entity.ChatId)

// _newFuncInitDestinations создает колбек для загрузки чатов (не используется)
// func _newFuncInitDestinations(s *Service) initDestinations {
// 	var fn initDestinations
// 	level := 0
// 	notFound := make(map[entity.ChatId]struct{})

// 	fn = func(destinations []entity.ChatId) {

// 		repeat := func() bool {
// 			var err error
// 			defer s.log.ErrorOrDebug(&err, "", "level", level)

// 			_, err = s.telegramRepo.LoadChats(&client.LoadChatsRequest{
// 				Limit: 200,
// 			})
// 			if err != nil {
// 				return false
// 			}
// 			for _, dstChatId := range destinations {
// 				_, err := s.telegramRepo.GetChatHistory(&client.GetChatHistoryRequest{
// 					ChatId:    dstChatId,
// 					Limit:     1,
// 					OnlyLocal: true,
// 				})
// 				if err != nil {
// 					notFound[dstChatId] = struct{}{}
// 					continue
// 				}
// 				delete(notFound, dstChatId)
// 			}
// 			if len(notFound) == 0 {
// 				return false
// 			}
// 			// TODO: было "level == 0", но рекурсия пока что отключена,
// 			// LoadChats() нельзя вызывать дважды, только если перезапускать клиент,
// 			// а это может привести к потере сообщений
// 			if level == 0 {
// 				a := []entity.ChatId{}
// 				for k := range notFound {
// 					a = append(a, k)
// 				}
// 				err = log.NewError("not found", "destinations", a)
// 				return false
// 			}
// 			level++
// 			return true
// 		}()
// 		if !repeat {
// 			return
// 		}

// 		fn(destinations) // !! хвостовая рекурсия
// 	}
// 	return fn
// }

// newFuncInitDestinations создает колбек для загрузки чатов
func newFuncInitDestinations(s *Service) initDestinations {
	loadChats := false
	return func(destinations []entity.ChatId) {
		if !loadChats {
			loadChats = true
			_, err := s.telegramRepo.LoadChats(&client.LoadChatsRequest{
				Limit: 200,
			})
			if err != nil {
				s.log.ErrorOrDebug(&err, "")
				return
			}
		}
		// Загружаем историю только для тех чатов, которые есть в конфигурации
		notFound := []entity.ChatId{}
		for _, dstChatId := range destinations {
			_, err := s.telegramRepo.GetChatHistory(&client.GetChatHistoryRequest{
				ChatId:    dstChatId,
				Limit:     1,
				OnlyLocal: true,
			})
			if err != nil {
				notFound = append(notFound, dstChatId)
			}
		}
		if len(notFound) > 0 {
			err := log.NewError("not found", "destinations", notFound)
			s.log.ErrorOrDebug(&err, "")
		}
	}
}
