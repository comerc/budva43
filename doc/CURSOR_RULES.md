# CURSOR RULES

Размер файлов `.cursor/rules/*.mdc`, который я могу эффективно воспринимать и использовать, зависит от нескольких факторов:

## Технические ограничения:

1. **Контекстное окно**: У меня есть ограничение на общий размер контекста в одной сессии (обычно около 200,000 токенов)

2. **Практические рекомендации**:
   - **Оптимально**: 1-5 KB (примерно 200-1000 строк текста)
   - **Хорошо**: до 10-15 KB 
   - **Максимум**: до 50 KB, но эффективность может снижаться

## Что влияет на восприятие:

- **Структурированность**: Хорошо организованный файл с заголовками и разделами воспринимается лучше
- **Конкретность**: Четкие, специфические правила работают лучше общих фраз
- **Приоритизация**: Важные правила лучше размещать в начале файла

## Рекомендуемая структура:

```markdown
# Основные принципы (самое важное)
# Архитектурные правила  
# Стандарты кодирования
# Документация
# Специфические инструкции
```

Для вашего Go проекта я бы рекомендовал файл размером **2-8 KB** - этого достаточно для всех ключевых правил, но не слишком много для эффективной обработки.
