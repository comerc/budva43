#!/bin/sh

# Скрипт для создания и применения coverage профайла в VS Code/Cursor

set -e

echo "🔍 Создаю coverage профайл..."

mkdir -p .coverage

# Создаем coverage профайл
GOEXPERIMENT=synctest go test -covermode=atomic -coverprofile=.coverage/.out -coverpkg=./... ./... > /dev/null || true

COVERAGE_EXCLUDE="(mocks|_easyjson\.go)"
grep -vE "$COVERAGE_EXCLUDE" .coverage/.out > .coverage/.txt
rm .coverage/.out

if [ -f ".coverage/.txt" ]; then
    COVERAGE=$(go tool cover -func=.coverage/.txt | tail -1 | awk '{print $NF}')
    echo ""
    echo "📊 Общее покрытие кода: $COVERAGE"
    echo ""
    echo "🎯 Для применения в Cursor:"
    echo "1. Нажмите Ctrl+Shift+P (Cmd+Shift+P на Mac)"
    echo "2. Введите 'Go: Apply Cover Profile'"
    echo "3. Укажите путь: $(pwd)/.coverage/.txt"
else
    echo "❌ Ошибка: .coverage/.txt не создан"
    exit 1
fi
