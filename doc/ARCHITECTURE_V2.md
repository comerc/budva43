# Соглашения по архитектуре проекта

## Принципы разделения слоёв и ответственности

### 1. Entity (Сущности)

**Назначение**:
- Представляют основные бизнес-сущности системы
- Содержат структуру данных и методы для работы с данными
- Выполняют функцию переноса данных между слоями (ранее функционал DTO)

**Правила**:
- Entity содержат структуры данных и методы для работы с ними
- Содержат поля и методы для трансформации и валидации данных
- Могут включать методы сериализации/десериализации
- Используются для передачи данных между слоями
- Для простых структур данных не используются функции-конструкторы
- Примеры: `ForwardRule`

**Пример**:
```go
// entity/entity.go
package entity

type ForwardRule struct {
    ID     string
    From   int64
    To     []int64
    Status RuleStatus
    // Другие поля
}

// Метод для сериализации объекта
func (r *ForwardRule) MarshalJSON() ([]byte, error) {
    // Реализация
}

// Метод для десериализации
func (r *ForwardRule) UnmarshalJSON(data []byte) error {
    // Реализация
}
```

### 2. Service (Сервисный слой)

**Назначение**:
- Содержит всю бизнес-логику приложения
- Реализует сценарии использования (use cases)
- Координирует работу между репозиториями и сущностями

**Правила**:
- Вся бизнес-логика должна быть в сервисном слое, а не в сущностях (entity)
- Сервисы принимают и возвращают entity
- Сервисы могут использовать репозитории для доступа к данным
- Применяются функции-конструкторы для внедрения зависимостей (через интерфейсы)
- Примеры: `MessageService`, `UserService`, `ForwardService`

**Пример**:
```go
// service/message/service.go
package message

type messageRepository interface {
  // Применяемые в этом модуле методы...    
}

type MessageService struct {
    repo messageRepository
}

func NewMessageService(repo messageRepository) *MessageService {
    return &MessageService{
        repo: repo,
    }
}

// GetText возвращает текст сообщения, если это текстовое сообщение
func (s *MessageService) GetText(message *entity.Message) string {
    if message == nil || message.Content == nil {
        return ""
    }
    
    if content, ok := message.Content.(*client.MessageText); ok {
        return content.Text.Text
    }
    
    return ""
}

// IsTextMessage проверяет, является ли сообщение текстовым
func (s *MessageService) IsTextMessage(message *entity.Message) bool {
    if message == nil || message.Content == nil {
        return false
    }
    
    _, ok := message.Content.(*client.MessageText)
    return ok
}
```

### 3. Repository (Репозитории)

**Назначение**:
- Предоставляют доступ к внешним системам и хранилищам данных
- Абстрагируют детали работы с API и базами данных
- Работают с Entity

**Правила**:
- Не должны содержать бизнес-логику
- Возвращают и принимают сущности (Entity)
- Используют функции-конструкторы для внедрения зависимостей (через интерфейсы)

**Пример**:
```go
// repository/telegram/message.go
package telegram

type client interface {
    // Применяемые в этом модуле методы...
}

type TelegramMessageRepository struct {
    client client
}

func NewTelegramMessageRepository(client client) *TelegramMessageRepository {
    return &TelegramMessageRepository{
        client: client    
    }
} 

func (r *TelegramMessageRepository) GetMessage(id int64) (*entity.Message, error) {
    // Реализация...
}
```

## Модифицированный подход к передаче данных

В связи с отказом от отдельного слоя DTO, функциональность передачи данных перешла к Entity:

1. **Бизнес-сущности как средство передачи данных**:
   - Entity используются для передачи данных между слоями
   - Entity могут содержать дополнительные поля и методы для преобразования форматов
   - Entity содержат методы для сериализации/десериализации
   - Entity обычно инициализируются с помощью литералов структур или фабричных методов в сервисах (но не с помощью функций-конструкторов)

2. **Когда можно добавлять методы к Entity**:
   - Методы для сериализации/десериализации (MarshalJSON, UnmarshalJSON)
   - Методы для преобразования в форматы для API (ToResponse)
   - Методы для валидации данных (Validate)
   - Вспомогательные методы для работы с данными (функции-получатели)

**Пример Entity с функциональностью передачи данных**:
```go
// entity/message.go
package entity

import (
    "encoding/json"
    "time"
)

type Message struct {
    ID         int64
    Text       string
    Date       time.Time
    SenderID   int64
    ChatID     int64
    MediaType  string
    MediaURL   string
}

// Метод для сериализации (ранее функционал DTO)
func (m *Message) MarshalJSON() ([]byte, error) {
    return json.Marshal(m)
}

// Метод для десериализации (ранее функционал DTO)
func (_ *Message) UnmarshalJSON(data []byte) (error) {
    var message Message
    err := json.Unmarshal(data, &message)
    return &message, err
}

// Метод для подготовки ответа API (ранее функционал DTO)
func (m *Message) ToResponse() map[string]interface{} {
    return map[string]interface{}{
        "id":         m.ID,
        "text":       m.Text,
        "date":       m.Date.Format(time.RFC3339),
        "sender_id":  m.SenderID,
        "chat_id":    m.ChatID,
        "media_type": m.MediaType,
        "media_url":  m.MediaURL,
    }
}
```

## Общие принципы

1. **Избегайте избыточных преобразований** - используйте прямую передачу Entity между слоями
2. **Разделяйте ответственность** - бизнес-логика только в сервисном слое
3. **Используйте интерфейсы** для абстрагирования внешних зависимостей
4. **Следуйте принципу DRY** - не дублируйте код и логику (без фанатизма)
5. **Минимизируйте избыточные абстракции** - не создавайте конструкторы для простых структур данных
6. **Тестируемость** - организуйте код так, чтобы его можно было легко тестировать

Придерживаясь этих соглашений, мы сможем создать чистую, поддерживаемую и тестируемую архитектуру, которая будет удобна для дальнейшего развития проекта.

