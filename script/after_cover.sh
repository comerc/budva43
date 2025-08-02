#!/bin/bash
set -e  # выход при любой ошибке
set -o pipefail  # выход при ошибке в любой команде пайплайна

if [ -f ".coverage/.txt" ]; then
    COVERAGE=$(go tool cover -func=.coverage/.txt | tail -1 | awk '{print $NF}')
    echo ""
    echo "📊 Общее покрытие кода: $COVERAGE"
    echo ""
    echo "🎯 Для применения в Cursor:"
    echo "1. Нажмите Ctrl+Shift+P (Cmd+Shift+P на Mac)"
    echo "2. Введите 'Go: Apply Cover Profile'"
    echo "3. Укажите путь: $(pwd)/.coverage/.txt"
    echo ""
else
    echo "❌ Ошибка: .coverage/.txt не создан"
    exit 1
fi
