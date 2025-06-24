#!/bin/bash

# Скрипт для создания и применения coverage профайла в VS Code/Cursor

set -e

echo "🔍 Создаю coverage профайл..."

# Создаем coverage профайл
GOEXPERIMENT=synctest go test -coverprofile=.coverage.out -coverpkg=./... ./test/ > /dev/null || true

if [ -f ".coverage.out" ]; then
    COVERAGE=$(go tool cover -func=.coverage.out | tail -1 | awk '{print $NF}')
    echo ""
    echo "📊 Общее покрытие кода: $COVERAGE"
    echo ""
    echo "🎯 Для применения в Cursor:"
    echo "1. Нажмите Cmd+Shift+P (или Ctrl+Shift+P на Linux/Windows)"
    echo "2. Введите 'Go: Apply Cover Profile'"
    echo "3. Укажите путь: $(pwd)/.coverage.out"
else
    echo "❌ Ошибка: .coverage.out не создан"
    exit 1
fi
