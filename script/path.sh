#!/bin/bash
set -e  # выход при любой ошибке
set -o pipefail  # выход при ошибке в любой команде пайплайна

echo "🔧 Исправление PATH для Go инструментов..."

# Проверяем переменную окружения
if [ -z "${GOPATH:-}" ]; then
    echo "ℹ️  Переменная \$GOPATH не установлена в текущей сессии"
fi

# Получаем GOPATH
GOPATH=$(go env GOPATH)
if [ -z "$GOPATH" ]; then
    echo "❌ GOPATH не установлен"
    exit 1
fi

echo "📁 GOPATH: $GOPATH"

# Определяем конфигурационный файл оболочки
SHELL_CONFIG=""
if [ -n "$ZSH_VERSION" ]; then
    SHELL_CONFIG="$HOME/.zshrc"
    echo "🐚 Обнаружена zsh"
elif [ -n "$BASH_VERSION" ]; then
    SHELL_CONFIG="$HOME/.bashrc"
    echo "🐚 Обнаружен bash"
else
    # Пробуем определить по умолчанию
    if [ -f "$HOME/.zshrc" ]; then
        SHELL_CONFIG="$HOME/.zshrc"
        echo "🐚 Найден .zshrc"
    elif [ -f "$HOME/.bashrc" ]; then
        SHELL_CONFIG="$HOME/.bashrc"
        echo "🐚 Найден .bashrc"
    fi
fi

if [ -z "$SHELL_CONFIG" ]; then
    echo "❌ Не удалось определить конфигурационный файл оболочки"
    echo "💡 Создайте файл ~/.bashrc или ~/.zshrc"
    exit 1
fi

echo "📄 Конфигурационный файл: $SHELL_CONFIG"

# Проверяем, есть ли уже GOPATH в PATH
if grep -q "GOPATH.*bin" "$SHELL_CONFIG" 2>/dev/null; then
    echo "✅ PATH уже настроен в $SHELL_CONFIG"
else
    echo "export PATH=\"\$PATH:$GOPATH/bin\"" >> "$SHELL_CONFIG"
    echo "✅ PATH добавлен в $SHELL_CONFIG"
fi

echo ""
echo "🔄 Применяем изменения..."
source "$SHELL_CONFIG"
echo "✅ PATH обновлен в текущей сессии"
echo "💡 Для других открытых терминалов выполните: source $SHELL_CONFIG"
echo ""
echo "📋 Установленные Go инструменты:"
ls -la "$GOPATH/bin" 2>/dev/null || echo "  (папка пуста или не существует)" 