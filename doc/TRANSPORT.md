# Транспортный слой

## Комбинация подходов для организации транспортного слоя

Наше предложение сочетает преимущества обоих подходов: сохранение контроллеров с разделением по бизнес-сущностям и выделение конкретных транспортных инструментов (роутеров) по протоколам.

Такая структура может выглядеть следующим образом:

```
budva43/
├── controller/           # Контроллеры по бизнес-сущностям
│   ├── message/          # Контроллер для работы с сообщениями
│   ├── forward/          # Контроллер для пересылки
│   └── report/           # Контроллер для отчетов
│
├── transport/            # Транспортные адаптеры по протоколам
│   ├── http/             # HTTP-роутеры и middleware
│   ├── bot/              # Обработка Telegram Bot API
│   └── cli/              # Интерфейс командной строки
```

## Взаимодействие между слоями

В этой архитектуре:

1. **Контроллеры (controller/)**:
   - Содержат бизнес-логику высокого уровня для конкретных сущностей
   - Вызывают методы сервисов и оркестрируют их работу
   - Независимы от транспортного слоя
   - Возвращают структурированные данные (не HTTP-ответы или Telegram-сообщения)

2. **Транспорт (transport/)**:
   - Содержит только код для работы с конкретными протоколами (HTTP, Telegram Bot API, CLI)
   - Вызывает соответствующие методы контроллеров
   - Преобразует входные данные из протокола в формат для контроллеров
   - Преобразует выходные данные контроллеров в формат протокола

## Пример реализации

```go
// controller/message/controller.go
package message

import (
    filterModel "github.com/example/budva43/model/filter"
    messageModel "github.com/example/budva43/model/message"
)

type messageService interface {
    // Применяемые в этом модуле методы...
}

type forwardService interface {
    // Применяемые в этом модуле методы...
}

type MessageController struct {
    messageService messageService
    forwardService forwardService
}

func NewMessageController(messageService messageService, forwardService forwardService) *MessageController {
    return &MessageController{
        messageService: messageService,
        forwardService: forwardService,
    }
}

// GetMessages возвращает список сообщений по фильтру
func (c *MessageController) GetMessages(filter *filterModel.Filter) ([]*messageModel.Message, error) {
    return c.messageService.GetMessages(filter)
}

// SendMessage отправляет новое сообщение
func (c *MessageController) SendMessage(message *messageModel.Message) (*messageModel.Message, error) {
    return c.messageService.SendMessage(message)
}

// transport/http/transport.go
package http

import (
    model "github.com/example/budva43/model/message"
    "net/http"
    "encoding/json"
)

type messageController interface {
    // Применяемые в этом модуле методы...
}

type forwardController interface {
    // Применяемые в этом модуле методы...    
}

type HTTPRouter struct {
    messageController messageController
    forwardController forwardController
}

func NewHTTPRouter(messageController messageController, forwardController forwardController) *HTTPRouter {
    return &HTTPRouter{
        messageController: messageController,
        forwardController: forwardController,
    }
}

// SetupRoutes настраивает HTTP-маршруты
func (r *HTTPRouter) SetupRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/api/messages", r.handleMessages)
    // Другие маршруты...
}

// handleMessages обрабатывает запросы на /api/messages
func (r *HTTPRouter) handleMessages(w http.ResponseWriter, req *http.Request) {
    switch req.Method {
    case http.MethodGet:
        // Извлечение параметров запроса в MessageFilter
        filter := parseFilterFromRequest(req)
        
        // Вызов контроллера
        messages, err := r.messageController.GetMessages(filter)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        // Преобразование в HTTP-ответ
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(messages)
        
    case http.MethodPost:
        // Декодирование JSON в Message
        var msg model.Message
        if err := json.NewDecoder(req.Body).Decode(&msg); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        // Вызов контроллера
        result, err := r.messageController.SendMessage(msg)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        // Преобразование в HTTP-ответ
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(result)
        
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

// Вспомогательная функция для извлечения фильтра из запроса
func parseFilterFromRequest(req *http.Request) model.MessageFilter {
    // Реализация...
}
```

## Преимущества данного подхода

1. **Чистое разделение ответственности**:
   - Контроллеры ответственны за бизнес-логику высокого уровня
   - Транспортный слой ответственен только за интеграцию с протоколами

2. **Удобная организация кода**:
   - Бизнес-логика сгруппирована по сущностям
   - Транспортные адаптеры сгруппированы по протоколам

3. **Упрощенное тестирование**:
   - Контроллеры можно тестировать независимо от транспортного слоя
   - Транспортный слой можно тестировать с использованием моков контроллеров

4. **Гибкость в разработке**:
   - Можно добавлять новые транспортные протоколы без изменения контроллеров
   - Можно добавлять новую функциональность в контроллеры без влияния на существующие транспортные адаптеры

5. **Плоская структура директорий**:
   - Отсутствие излишней вложенности директорий
   - Простая навигация по проекту

## Потенциальные недостатки и их решения

1. **Дублирование маппинга между URL и контроллерами**:
   - Решение: использовать декларативное определение маршрутов или генерацию на основе аннотаций/комментариев

2. **Размытие границы между контроллерами и сервисами**:
   - Решение: четко определить ответственность каждого слоя и следовать этим определениям

## Вывод

Предложенный гибридный подход представляет собой разумный компромисс, сохраняя преимущества группировки по бизнес-сущностям для контроллеров и по протоколам для транспортного слоя. Такая структура обеспечивает хороший баланс между организацией кода и простотой навигации.

Этот подход хорошо подходит для средних и крупных проектов, где есть несколько бизнес-сущностей и несколько транспортных протоколов, что соответствует описанию проекта Budva43.
