# Отображение подсказки пароля (Password Hint)

## Обзор

При авторизации в Telegram через двухфакторную аутентификацию (2FA) приложение может отображать подсказку пароля, которая помогает пользователю вспомнить свой пароль.

## Как это работает

### 1. Структура данных

TDLib предоставляет состояние `AuthorizationStateWaitPassword` с полем `PasswordHint`:

```go
type AuthorizationStateWaitPassword struct {
    // Подсказка для пароля; может быть пустой
    PasswordHint string `json:"password_hint"`
    // Другие поля...
}
```

### 2. Терминальный интерфейс

В терминальном интерфейсе (`transport/term/transport.go`):

```go
case client.TypeAuthorizationStateWaitPassword:
    passwordState := state.(*client.AuthorizationStateWaitPassword)
    if passwordState.PasswordHint != "" {
        t.termRepo.Printf("Введите пароль (подсказка: %s): \n", passwordState.PasswordHint)
    } else {
        t.termRepo.Println("Введите пароль: ")
    }
```

**Пример вывода:**
- С подсказкой: `Введите пароль (подсказка: my secret hint): `
- Без подсказки: `Введите пароль: `

### 3. Веб-интерфейс

В веб-интерфейсе (`transport/web/handler_rest.go`) подсказка возвращается в JSON-ответе:

```go
// GET /api/auth/telegram/state
{
    "state_type": "authorizationStateWaitPassword",
    "password_hint": "my secret hint"  // только если подсказка не пустая
}
```

**Примеры ответов:**

С подсказкой:
```json
{
    "state_type": "authorizationStateWaitPassword",
    "password_hint": "my secret hint"
}
```

Без подсказки:
```json
{
    "state_type": "authorizationStateWaitPassword"
}
```

## Использование

### Makefile

Для тестирования через curl:

```bash
make test-auth-telegram-state
make test-auth-telegram-submit-password
```

### Интеграция с фронтендом

Фронтенд может проверить наличие поля `password_hint` в ответе и отобразить подсказку пользователю:

```javascript
const response = await fetch('/api/auth/telegram/state');
const data = await response.json();

if (data.state_type === 'authorizationStateWaitPassword') {
    const hint = data.password_hint;
    if (hint) {
        // Отобразить подсказку пользователю
        showPasswordInput(`Введите пароль (подсказка: ${hint})`);
    } else {
        showPasswordInput('Введите пароль');
    }
}
```

## Безопасность

- Подсказка пароля передается от Telegram и не содержит сам пароль
- Подсказка помогает пользователю вспомнить пароль, но не раскрывает его
- Если подсказка пустая, она не включается в ответ API 