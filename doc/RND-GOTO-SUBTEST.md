# RND: Навигация к case табличных тестов Go в VSCode

## Проблема
- В Go табличные тесты реализуются через цикл с t.Run(caseName, ...), где caseName часто формируется динамически.
- Стандартные средства VSCode (Go extension + Test Explorer) отображают имена subtest-ов, но переходят только к началу функции-обёртки, а не к конкретному case.
- gopls не сопоставляет имя case-а с позицией в исходнике, поэтому штатная навигация невозможна.

## Анализ существующих решений
- [exp-vscode-go (Go Companion)](https://github.com/firelizzard18/exp-vscode-go): позволяет расширять Test Explorer и добавлять свои действия в контекстное меню тестов.
- [vscode-go-tdt-outline](https://github.com/toga4/vscode-go-tdt-outline): парсит табличные тесты, отображает case-ы в Outline, позволяет переходить к началу case-а (позиция в исходнике определяется парсером, а не gopls).
- Оба решения не интегрированы между собой и не дают полного UX: Test Explorer не умеет переходить к case, а Outline не связан с запуском тестов.

## Возможные пути реализации
1. **Собственное расширение VSCode**
   - Парсить исходники Go для поиска табличных тестов и case-ов (анализировать t.Run внутри циклов).
   - Добавлять case-ы как отдельные элементы в Test Explorer с возможностью перехода к их позиции в исходнике.
   - (Опционально) Связывать запуск конкретного case-а с запуском теста с фильтром по имени (go test -run ...).
   - Возможна интеграция с exp-vscode-go для расширения Test Explorer.
2. **Доработка существующих решений**
   - Внести вклад в vscode-go-tdt-outline: добавить интеграцию с Test Explorer или запуск тестов по case-у.
   - Внести вклад в exp-vscode-go: реализовать парсинг case-ов и навигацию к ним.
3. **Ограничения**
   - Если caseName формируется не строкой, а выражением, точная позиция может быть не определена.
   - Для корректной навигации желательно, чтобы caseName был явно задан строкой.

## Рекомендации для будущей реализации
- Изучить API Test Explorer и возможности расширения его структуры.
- Изучить опыт реализации парсинга case-ов в vscode-go-tdt-outline.
- Прототипировать парсер табличных тестов на Go/TS для извлечения case-ов и их позиций.
- Оценить возможность автоматического сопоставления case-ов с запуском go test -run ^TestFunc$/^caseName$.
- Рассмотреть вариант объединения Outline и Test Explorer для единого UX.

---

**Этот файл фиксирует только контекст и направления для будущей реализации переходов к case-ам табличных тестов Go в VSCode.** 