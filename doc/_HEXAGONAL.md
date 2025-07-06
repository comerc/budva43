# Применение гексагональной архитектуры (черновик)

## Анализ текущей архитектуры и проблемы

Изучив документы и код, я вижу следующую картину:

### Текущая архитектура
Проект использует **луковичную архитектуру** с четким разделением на слои:
- **Entity** - бизнес-сущности
- **Service** - бизнес-логика  
- **Repo** - доступ к данным
- **Transport** - транспортные адаптеры

### Проблема с serviceEngine
`serviceEngine` действительно разросся и стал выполнять слишком много функций:
1. **Валидация конфигурации** (`validateConfig`)
2. **Обогащение конфигурации** (`enrichConfig`) 
3. **Управление жизненным циклом** (`Start`, `run`, `Close`)
4. **Диспетчеризация обновлений** (`handleUpdates`)

Для решения этой проблемы была проведена декомпозиция - выделены отдельные **handlers** для каждого типа обновлений.

### Handlers в текущей архитектуре
Handlers сейчас находятся "между" слоями:
- Они не являются чистыми сервисами (содержат специфичную логику обработки Telegram-обновлений)
- Они не являются транспортными адаптерами (не работают с протоколами)
- Они оркеструют работу множества сервисов для обработки конкретных событий

## Возможность применения гексагональной архитектуры

**Да, применение гексагональной архитектуры для handlers возможно и даже желательно!** Вот почему:

### 1. Handlers как Application Services (порты)

В гексагональной архитектуре handlers могут стать **Application Services** - портами, которые:
- Определяют интерфейсы для входящих операций (inbound ports)
- Оркеструют работу доменных сервисов
- Не содержат бизнес-логики, а только координируют её выполнение

```go
// Порт для обработки входящих сообщений
type NewMessageHandler interface {
    Handle(ctx context.Context, update *client.UpdateNewMessage) error
}

// Порт для обработки редактирования сообщений  
type MessageEditedHandler interface {
    Handle(update *client.UpdateMessageEdited) error
}
```

### 2. ServiceEngine как Adapter (адаптер)

`serviceEngine` становится **входящим адаптером** (inbound adapter), который:
- Получает события от Telegram API
- Маршрутизирует их к соответствующим портам (handlers)
- Управляет жизненным циклом

```go
// Входящий адаптер для Telegram
type TelegramAdapter struct {
    newMessageHandler     NewMessageHandler
    messageEditedHandler  MessageEditedHandler
    // другие handlers
}

func (a *TelegramAdapter) handleUpdates(listener *client.Listener) {
    for update := range listener.Updates {
        switch updateByType := update.(type) {
        case *client.UpdateNewMessage:
            a.newMessageHandler.Handle(ctx, updateByType)
        case *client.UpdateMessageEdited:
            a.messageEditedHandler.Handle(updateByType)
        }
    }
}
```

### 3. Структура гибридной архитектуры

```
app/
├── domain/              # Доменный слой (луковичная архитектура)
│   ├── entity/         # Бизнес-сущности
│   ├── service/        # Доменные сервисы
│   └── repo/           # Интерфейсы репозиториев
│
├── application/        # Слой приложения (гексагональная архитектура)
│   ├── port/           # Порты (интерфейсы)
│   │   ├── inbound/    # Входящие порты (handlers)
│   │   └── outbound/   # Исходящие порты (repo interfaces)
│   └── service/        # Application Services (реализации портов)
│       ├── message_handler.go
│       ├── edit_handler.go
│       └── delete_handler.go
│
├── adapter/            # Адаптеры (гексагональная архитектура)
│   ├── inbound/        # Входящие адаптеры
│   │   ├── telegram/   # TelegramAdapter (бывший serviceEngine)
│   │   ├── web/        # HTTP адаптер
│   │   └── term/       # term адаптер
│   └── outbound/       # Исходящие адаптеры
│       ├── storage/    # StorageAdapter
│       └── telegram/   # TelegramRepoAdapter
│
└── config/             # Конфигурация
```

### 4. Преимущества гибридного подхода

**✅ Сохраняются преимущества луковичной архитектуры:**
- Четкое разделение доменной логики
- Инверсия зависимостей
- Тестируемость доменных сервисов

**✅ Добавляются преимущества гексагональной архитектуры:**
- Четкое разделение портов и адаптеров
- Легкость добавления новых способов взаимодействия
- Изоляция от внешних технологий

**✅ Решается проблема с handlers:**
- Handlers становятся Application Services с четкой ответственностью
- ServiceEngine упрощается до роли адаптера
- Улучшается тестируемость

### 5. Пример реализации

```go
// application/port/inbound/message_handler.go
package inbound

type NewMessageHandler interface {
    Handle(ctx context.Context, update *client.UpdateNewMessage) error
}

// application/service/message_handler.go  
package service

type NewMessageHandlerService struct {
    messageService    domain.MessageService
    forwarderService  domain.ForwarderService
    storageService    domain.StorageService
    // другие доменные сервисы
}

func (h *NewMessageHandlerService) Handle(ctx context.Context, update *client.UpdateNewMessage) error {
    // Оркестрация доменных сервисов без бизнес-логики
    message := h.messageService.ProcessUpdate(update)
    rules := h.forwarderService.GetRules(message.ChatId)
    return h.forwarderService.Forward(message, rules)
}

// adapter/inbound/telegram/adapter.go
package telegram

type Adapter struct {
    newMessageHandler inbound.NewMessageHandler
    // другие handlers
}

func (a *Adapter) handleUpdates(listener *client.Listener) {
    // Простая маршрутизация без бизнес-логики
}
```

## Рекомендация

**Да, применение гексагональной архитектуры для handlers не только возможно, но и рекомендуется!**

Это позволит:
1. **Упростить serviceEngine** до роли простого адаптера
2. **Структурировать handlers** как Application Services с четкой ответственностью  
3. **Сохранить** все преимущества текущей луковичной архитектуры
4. **Улучшить** тестируемость и расширяемость системы
5. **Не переусложнить** код - изменения будут локальными и логичными

Гибридный подход (луковичная + гексагональная) является естественным развитием архитектуры и хорошо решает проблему разросшегося serviceEngine.

-----

**НЕ применяйте гексагональную архитектуру для handlers!**

Причины:
1. **Текущая структура работает** - handlers хорошо декомпозированы
2. **Нет реальной проблемы** - serviceEngine уже упрощен до диспетчера
3. **Переусложнение** - гексагональная архитектура добавит сложности без пользы
4. **Когнитивная нагрузка** - потребует постоянного разделения application/domain logic

**Лучшее решение**: оставить текущую луковичную архитектуру и сделать минимальные улучшения в serviceEngine для упрощения диспетчеризации.

Иногда **не делать ничего** - лучшее архитектурное решение! 🎯