# Соглашения по архитектуре проекта

## Принципы разделения слоёв и ответственности

### 1. Entity (Сущности)

**Назначение**:
- Представляют основные бизнес-сущности системы
- Содержат структуру данных и минимальный набор методов
- Встраивают структуры из внешних библиотек (go-tdlib)

**Правила**:
- Entity должны быть максимально простыми структурами данных
- Содержат только необходимые поля и конструкторы
- Могут содержать простые геттеры/сеттеры, но без сложной бизнес-логики
- Должны отражать модель данных, но не бизнес-процессы
- Примеры: `Message`, `Chat`, `User`, `ForwardRule`

**Пример**:
```go
// entity/message/entity.go
package message

type Message struct {
    // Встраиваем структуру из TDLib
    *client.Message
    
    // Дополнительные поля
    ParsedDate time.Time
}

func NewMessage(tdlibMessage *client.Message) *Message {
    if tdlibMessage == nil {
        return nil
    }
    
    return &Message{
        Message:    tdlibMessage,
        ParsedDate: time.Unix(int64(tdlibMessage.Date), 0),
    }
}
```

### 2. Service (Сервисный слой)

**Назначение**:
- Содержит всю бизнес-логику приложения
- Реализует сценарии использования (use cases)
- Координирует работу между репозиториями и сущностями

**Правила**:
- Вся бизнес-логика должна быть в сервисном слое, а не в сущностях
- Сервисы принимают Entity как параметры и работают с ними
- Сервисы могут использовать репозитории для доступа к данным
- Сервисы могут преобразовывать сущности в DTO при необходимости
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

**Пример**:
```go
// repository/telegram/message.go
package telegram

type TelegramMessageRepository struct {
    client *client.Client
}

func (r *TelegramMessageRepository) GetMessage(id int64) (*entity.Message, error) {
    // Реализация
}
```

### 4. DTO (Data Transfer Objects)

**Назначение**:
- Используются только для передачи данных между слоями или системами
- Применяются в случаях, когда формат данных должен отличаться от Entity
- Часто используются для API, UI или интеграций

**Правила**:
- DTO должны быть простыми структурами без бизнес-логики
- Используются, только когда действительно нужно преобразование форматов
- Преобразование между Entity и DTO должно происходить в сервисном слое

## Когда использовать DTO

DTO следует использовать в следующих случаях:

1. **Публичное API**:
   - Когда требуется предоставить API, где формат данных отличается от внутренних сущностей
   - Для скрытия внутренней структуры данных от внешних потребителей

2. **Агрегирование данных**:
   - Когда необходимо объединить данные из нескольких сущностей в один объект
   - Для представления сложного вида, собранного из разных источников

3. **Частичное представление**:
   - Когда требуется вернуть только часть полей сущности
   - Для оптимизации передачи данных

4. **Преобразование форматов**:
   - Когда внешняя система требует данные в формате, отличном от внутреннего представления
   - Для адаптации данных под требования интеграции

**Пример DTO для API**:
```go
// model/message/model.go
package message

type MessageResponse struct {
    ID         int64     `json:"id"`
    Text       string    `json:"text"`
    Date       time.Time `json:"date"`
    SenderName string    `json:"sender_name"`
    ChatName   string    `json:"chat_name"`
    MediaType  string    `json:"media_type,omitempty"`
}

// В сервисе
func (s *MessageService) GetMessageForAPI(messageID int64) (*model.MessageResponse, error) {
    message, err := s.repo.GetMessage(messageID)
    if err != nil {
        return nil, err
    }
    
    // Получаем дополнительные данные
    sender, err := s.userRepo.GetUser(message.SenderUserID)
    if err != nil {
        return nil, err
    }
    
    chat, err := s.chatRepo.GetChat(message.ChatID)
    if err != nil {
        return nil, err
    }
    
    // Конвертируем в DTO
    return &model.MessageResponse{
        ID:         message.ID,
        Text:       s.GetText(message),
        Date:       message.ParsedDate,
        SenderName: sender.GetFullName(),
        ChatName:   chat.Title,
        MediaType:  s.GetContentType(message),
    }, nil
}
```

## Общие принципы

1. **Избегайте дублирования** - не создавайте DTO, если они почти идентичны Entity
2. **Разделяйте ответственность** - бизнес-логика только в сервисном слое
3. **Используйте интерфейсы** для абстрагирования внешних зависимостей
4. **Следуйте принципу DRY** - не дублируйте код и логику (без фанатизма)
5. **Тестируемость** - организуйте код так, чтобы его можно было легко тестировать

Придерживаясь этих соглашений, мы сможем создать чистую, поддерживаемую и тестируемую архитектуру, которая будет удобна для дальнейшего развития проекта.

