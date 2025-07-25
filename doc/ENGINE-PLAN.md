# План реализации механизма форвардинга сообщений

## Основной поток обработки сообщений

Изучив архитектуру проекта Budva32 и документацию Budva43, я вижу, что основной поток обработки сообщений должен быть реализован следующим образом:

### 1. Структура конфигурации

В файле `config/config.go` нужно расширить структуру `forwarding` для включения настроек из Budva32:

```go
// Настройки движка форвардинга из budva32
engine struct {
  // Настройки для замены ссылок на себя
  ReplaceMyselfLinks map[int64]domain.ReplaceMyselfLinkSettings
  // Настройки для замены фрагментов текста
  ReplaceFragments map[int64]domain.ReplaceFragmentSettings
  // Настройки источников
  Sources map[int64]domain.Source
  // Настройки отчетов
  Reports struct {
    Template string
    For      []int64
  }
  // Правила форвардинга
  Forwards map[string]domain.ForwardRule
  // Настройки автоответов
  Answers map[int64]domain.Answer
  // Удаление системных сообщений
  DeleteSystemMessages map[int64]bool
}
```

### 2. Реализация сервисов

1. **MessageService** - основной сервис для работы с сообщениями
   - Обрабатывает получение и отправку сообщений
   - Методы для работы с текстом, фото, медиа и другими типами сообщений

2. **ForwardService** - сервис для пересылки сообщений
   - Реализует логику пересылки и копирования сообщений
   - Обрабатывает правила пересылки из конфигурации
   - Синхронизирует редактирование и удаление сообщений

3. **TransformService** - сервис для преобразования текста
   - Замена фрагментов текста
   - Обработка ссылок и упоминаний

4. **FilterService** - сервис для фильтрации сообщений
   - Проверка сообщений на соответствие правилам включения/исключения
   - Обработка регулярных выражений

5. **HistoryService** - сервис для отслеживания истории сообщений
   - Хранение связей между оригинальными и пересланными сообщениями
   - Поддержка редактирования и удаления связанных сообщений

### 3. Основной поток обработки сообщений

Основной поток обработки сообщений должен быть реализован в следующих файлах:

1. **service/engine/service.go** - сервис для работы с Telegram API
   - Подписка на обновления Telegram
   - Обработка различных типов обновлений (новые сообщения, редактирование, удаление)
   - Направление обновлений в соответствующие обработчики

2. **service/message/service.go** - сервис для обработки сообщений
   - Получение текста из разных типов сообщений
   - Работа с медиафайлами
   - Отправка сообщений

3. **service/forward/service.go** - сервис для пересылки
   - Реализация алгоритма форвардинга
   - Обработка правил фильтрации
   - Синхронизация редактирования и удаления

В файле `main.go` должны быть:
- Инициализация зависимостей
- Запуск обработчика сообщений Telegram
- Настройка конфигурации и хранилища

### 4. Обработка типов обновлений

1. **UpdateNewMessage**
   - Проверка правил форвардинга
   - Фильтрация по правилам включения/исключения
   - Отправка копий или пересылка в целевые чаты

2. **UpdateMessageEdited**
   - Нахождение связанных копий сообщения
   - Синхронное редактирование копий
   - Применение правил трансформации текста

3. **UpdateDeleteMessages**
   - Нахождение связанных копий сообщения
   - Удаление копий (если не установлен флаг Indelible)

4. **UpdateMessageSendSucceeded**
   - Сохранение отношений между временными и постоянными Id сообщений

### 5. Хранение отношений между сообщениями

В **StorageService** необходимо реализовать методы для работы с хранилищем BadgerDB:
- Сохранение отношений между исходными и скопированными сообщениями
- Получение списка копий для исходного сообщения
- Сохранение временных и постоянных Id сообщений

### 6. Обработка медиа-альбомов

Для работы с медиа-альбомами (группами фото/видео) необходимо:
- Накапливать сообщения альбома до полного его получения
- Пересылать/копировать сразу весь альбом
- Обрабатывать подписи к первому сообщению альбома

### Заключение

Основная логика форвардинга должна быть реализована в сервисах с чёткими зонами ответственности:
- **engine/service.go** - прием обновлений от Telegram
- **message/service.go** - обработка различных типов сообщений
- **forward/service.go** - основная логика пересылки
- **filter/service.go** - фильтрация сообщений
- **transform/service.go** - преобразование текста и форматирования

Следуя принципам чистой архитектуры, данный функционал должен быть разделен на отдельные модули, что облегчит тестирование и дальнейшее развитие.
