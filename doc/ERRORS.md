# Соглашение по обработке ошибок

![](https://habrastorage.org/webt/os/mi/cg/osmicg2qjlfyz5spftc0dy_4qvk.jpeg)

Go не признаёт сложных развесистых фреймворков — и это правильно. Любой инструмент должен быть настолько очевидным, чтобы разработчики сразу видели пользу от его применения.

И если вы сталкивались на больших legacy-проектах с хаосом в обработке и логировании ошибок — когда одна и та же ошибка дублируется в логах на каждом уровне, когда непонятно где обрабатывать ошибку, а где передавать дальше, когда тесты превращаются в кромешный ад из-за необходимости мокать каждый возможный путь ошибки — тогда возможно пригодится то, что я хочу предложить.

## Проблема error-hell

Go славится своей прямолинейностью в обработке ошибок (при этом отпугивая новичков). Но в реальных приложениях такая простота оборачивается избыточностью, которая часто засоряет код и лишь усложняет сопровождение.

Представьте: вы пишете обработку списка элементов в цикле. Один или несколько элементов не удалось обработать — что делать? По канонам Go нужно накопить ошибки в цикле и вернуть их вверх. Но зачем, если вы хотите всего лишь продолжить обработку остальных элементов? Ошибка не влияет на логику программы, но заставляет писать лишний код передачи ошибок и дублировать логирование на каждом уровне.

Или горутины: как элегантно обработать ошибку в фоновой задаче? Канал ошибок? Контекст? А если ошибка не критична и нужна только для отладки?

А ещё игра в испорченный телефон, как следствие от повторных обработок ошибок: `return fmt.Errorf("level1: %w", err)`, `return fmt.Errorf("level2: %w", err)`, `return fmt.Errorf("level3: %w", err)`. В итоге получаем многословные цепочки, которые дублируют путь ошибки вверх по стеку.

И главное — **возврат ошибок служит двум целям**: управление ветвлением программы и обеспечение тестируемости. Но что если эти цели можно разделить? Структурированное логирование с инструментами вроде [comerc/spylog](https://github.com/comerc/spylog) решает проблему тестирования без необходимости возвращать ошибки вверх по стеку.

**Решение:** минимальная надстройка в соглашениях, которая сохраняет дух Go — "явное лучше скрытого" и "обрабатывай ошибки по месту", но убирает механическую избыточность.

## Принципы

1. **Обрабатывай ошибку по месту** - если ошибка не влияет на ветвление программы
2. **Передавай ошибку вверх** - только если нужно изменить ход выполнения программы
3. **Не дублируй логи** - либо логируй ошибку, либо передавай её вверх

## Паттерны

### ✅ Обработка по месту (не влияет на ветвление)

```go
// В циклах
for _, item := range items {
    if err := processItem(item); err != nil {
        log.Error("failed to process item", "item", item.ID, "error", err)
        continue // продолжаем работу
    }
}

// В горутинах
go func() {
    if err := backgroundTask(); err != nil {
        log.Error("background task failed", "error", err)
        // не возвращаем ошибку - обработали по месту
    }
}()
```

### ✅ Передача вверх (влияет на ветвление)

```go
func validateUser(id string) error {
    user, err := db.GetUser(id)
    if err != nil {
        return err // передаём как есть, без обёртки
    }

    if user.Status != "active" {
        return ErrUserInactive // возвращаем типизированную ошибку
    }

    if user.Info == nil {
        return fmt.Errorf("%w", &ErrUserInfo{data: "data"})
        // уточняем ошибку - добавляем data
    }

    return nil
}

// Использование
func handleRequest(userID string) {
    if err := validateUser(userID); err != nil {
        if errors.Is(err, ErrUserInactive) {
            // специальная обработка
        }
        var userInfoErr *ErrUserInfo
        if errors.As(err, &userInfoErr) {
            // обработка для ErrUserInfo
        }
        log.Error("validation failed", "error", err)
        // логируем ошибку, только когда не передаём выше по стеку
        return
    }
}
```

### ❌ Избегаем обёртывания без цели

```go
// Плохо - создаём информационный шум
func getConfig() (*Config, error) {
    data, err := os.ReadFile("config.json")
    if err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err) // лишняя обёртка
    }
    // ...
}

// Хорошо - передаём как есть
func getConfig() (*Config, error) {
    data, err := os.ReadFile("config.json")
    if err != nil {
        return nil, err // передаём оригинальную ошибку
    }
    // ...
}
```

## Типизированные ошибки для ветвления

```go
var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserInactive = errors.New("user inactive")
    ErrInvalidInput = errors.New("invalid input")
)

// Проверка типа ошибки
if errors.Is(err, ErrUserNotFound) {
    // специальная обработка
}
```

## Структурированные логи вместо возврата ошибок

```go
// Вместо возврата ошибки из горутины
func processInBackground(items []Item) {
    for _, item := range items {
        if err := process(item); err != nil {
            log.Error("item processing failed",
                "item_id", item.ID,
                "error", err,
                "retry_count", item.RetryCount)
            // продолжаем обработку других элементов
        }
    }
}
```

## Правило принятия решения

**Спроси себя: "Изменит ли эта ошибка ход выполнения программы?"**

- **Да** → передай ошибку вверх, типизируя по необходимости
- **Нет** → залогируй и продолжай выполнение

Это соглашение сохраняет простоту Go и убирает избыточность без введения сложных абстракций.

## Тестирование с применением структурированных логов

### Зачем spylog?

Когда мы обрабатываем ошибки по месту через логирование, возникает вопрос: как тестировать такой код? Вместо проверки возвращённой ошибки нам нужен способ убедиться, что логирование действительно произошло с правильными данными.

Библиотека [comerc/spylog](https://github.com/comerc/spylog) перехватывает записи в лог для конкретного модуля в рамках теста, позволяя проверить:
- Случилось ли логирование
- Правильное ли сообщение
- Корректные ли атрибуты (включая тип ошибки)

### Пример с spylog

```go
// Код модуля
type UserService struct {
    log *slog.Logger
}

func NewUserService() *UserService {
    return &UserService{
        log: log.NewLogger(),
    }
}

func (s *UserService) ProcessUsers(users []User) {
    for _, user := range users {
        if err := s.validateUser(user); err != nil {
            // Логируем ошибку по месту, не возвращаем вверх
            s.log.Error("user validation failed",
                "user_id", user.ID,
                "email", user.Email,
                "error", err,
            )
            continue // продолжаем обработку других пользователей
        }
        // обрабатываем валидного пользователя
    }
}

// Тест
func TestProcessUsers(t *testing.T) {
    var service *UserService
    logHandler := spylog.GetHandler(t.Name(), func() {
        service = NewUserService()
    })

    users := []User{
        {ID: "1", Email: "valid@example.com"},
        {ID: "2", Email: "invalid-email"}, // невалидный
    }

    service.ProcessUsers(users)

    // Проверяем что залогирована одна ошибка
    require.Len(t, logHandler.Records, 1)
    record := logHandler.Records[0]

    assert.Equal(t, "user validation failed", record.Message)
    assert.Equal(t, "2", spylog.GetAttrValue(record, "user_id"))
    assert.Equal(t, "invalid-email", spylog.GetAttrValue(record, "email"))
    assert.Equal(t, ErrInvalidEmail, errors.Is(spylog.GetAttrValue(record, "error"))
}
```

### Преимущества подхода

- **Разделение ответственности**: ошибки для ветвления vs ошибки для отладки
- **Простота тестирования**: проверяем логи вместо возвращаемых ошибок
- **Меньше шаблонного кода**: не нужно передавать ошибки вверх по стеку

P.S. проблема тестирования асинхронного кода так же решается с применением этого подхода через другую надстройку - [synctest-experiment](https://danp.net/posts/synctest-experiment/).

---

[опубликовано](https://habr.com/ru/articles/912150/)
