#!/bin/bash
set -e  # выход при любой ошибке
set -o pipefail  # выход при ошибке в любой команде пайплайна

echo "🚀 Инициализация проекта budva43..."

# Определяем ОС
OS=$(uname -s)
case "$OS" in
    Linux*)     PLATFORM="linux";;
    Darwin*)    PLATFORM="macos";;
    CYGWIN*)    PLATFORM="windows";;
    MINGW*)     PLATFORM="windows";;
    *)          echo "❌ Неподдерживаемая ОС: $OS"; exit 1;;
esac

echo "📋 Платформа: $PLATFORM"

# Функция для установки Go инструментов
install_go_tools() {
    echo "🔧 Установка Go инструментов..."
    
    # Основные инструменты
    go install github.com/vektra/mockery/v2@v2.53.3
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
    # go install github.com/mailru/easyjson/...@latest
    go install github.com/99designs/gqlgen@latest
    go install github.com/unused-interface-methods/unused-interface-methods@latest
    go install github.com/error-log-or-return/error-log-or-return@latest
    go install github.com/go-task/task/v3/cmd/task@latest
    go install github.com/michurin/human-readable-json-logging/cmd/...@latest

    # BDD тестирование
    go install github.com/cucumber/godog/cmd/godog@latest
    
    # Protocol Buffers
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    
    # gRPC инструменты
    go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
    
    # JSON инструменты
    go install github.com/itchyny/gojq/cmd/gojq@latest
    go install github.com/noahgorstein/jqp@latest
    
    echo "✅ Go инструменты установлены"
}

# Функция для установки системных зависимостей на Ubuntu/Debian
install_linux_deps() {
    echo "📦 Установка системных зависимостей (Ubuntu/Debian)..."
    
    sudo apt-get update
    sudo apt-get install -y \
        git \
        curl \
        wget \
        build-essential \
        pkg-config \
        libssl-dev \
        zlib1g-dev \
        libstdc++6 \
        bash \
        tar \
        lsof \
        ffmpeg \
        yt-dlp \
        protobuf-compiler
    
    echo "✅ Системные зависимости установлены"
}

# Функция для установки системных зависимостей на macOS
install_macos_deps() {
    echo "📦 Установка системных зависимостей (macOS)..."
    
    # Проверяем Homebrew
    if ! command -v brew >/dev/null 2>&1; then
        echo "❌ Homebrew не установлен. Установите: https://brew.sh"
        exit 1
    fi
    
    brew install \
        git \
        curl \
        wget \
        bash \
        lsof \
        ffmpeg \
        yt-dlp \
        git-filter-repo \
        protobuf \
        grpcurl
    
    echo "✅ Системные зависимости установлены"
}

# Функция для установки системных зависимостей на Windows
install_windows_deps() {
    echo "📦 Установка системных зависимостей (Windows)..."
    
    # Список зависимостей для Windows
    WINDOWS_DEPS="git ffmpeg yt-dlp protobuf"
    
    # Проверяем наличие пакетных менеджеров
    HAS_CHOCO=false
    HAS_SCOOP=false
    
    if command -v choco >/dev/null 2>&1; then
        HAS_CHOCO=true
        echo "✅ Найден Chocolatey"
    fi
    
    if command -v scoop >/dev/null 2>&1; then
        HAS_SCOOP=true
        echo "✅ Найден Scoop"
    fi
    
    # Если оба менеджера установлены, предлагаем выбор
    if [ "$HAS_CHOCO" = true ] && [ "$HAS_SCOOP" = true ]; then
        echo ""
        echo "🤔 Найдены оба пакетных менеджера. Выберите один:"
        echo "1) Chocolatey (рекомендуется)"
        echo "2) Scoop"
        echo ""
        read -p "Введите номер (1 или 2): " -n 1 -r
        echo
        
        if [[ $REPLY =~ ^[1]$ ]]; then
            echo "📦 Установка зависимостей через Chocolatey..."
            choco install -y $WINDOWS_DEPS
            echo "✅ Системные зависимости установлены через Chocolatey"
        elif [[ $REPLY =~ ^[2]$ ]]; then
            echo "📦 Установка зависимостей через Scoop..."
            scoop install $WINDOWS_DEPS
            echo "✅ Системные зависимости установлены через Scoop"
        else
            echo "❌ Неверный выбор. Завершение работы."
            exit 1
        fi
    elif [ "$HAS_CHOCO" = true ]; then
        echo "📦 Установка зависимостей через Chocolatey..."
        choco install -y $WINDOWS_DEPS
        echo "✅ Системные зависимости установлены через Chocolatey"
    elif [ "$HAS_SCOOP" = true ]; then
        echo "📦 Установка зависимостей через Scoop..."
        scoop install $WINDOWS_DEPS
        echo "✅ Системные зависимости установлены через Scoop"
    else
        echo "❌ Не найден ни один пакетный менеджер"
        echo "📥 Установите один из пакетных менеджеров:"
        echo ""
        echo "Chocolatey (рекомендуется):"
        echo "  Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))"
        echo ""
        echo "Scoop:"
        echo "  Set-ExecutionPolicy RemoteSigned -Scope CurrentUser"
        echo "  irm get.scoop.sh | iex"
        echo ""
        echo "После установки пакетного менеджера запустите скрипт снова"
        exit 1
    fi
}

# Функция для создания конфигурационных файлов
setup_config() {
    echo "🔧 Настройка конфигурации..."
    
    # Создаем .env файл если его нет
    if [ ! -f .config/.private/.env ]; then
        mkdir -p .config/.private
        touch .config/.private/.env
        echo "✅ Создан пустой .env файл"
    fi
    
    echo "✅ Конфигурация настроена"
}

# Функция для проверки Go
check_go() {
    if ! command -v go >/dev/null 2>&1; then
        echo "❌ Go не установлен"
        echo "📥 Рекомендуем установить через g (Golang Version Manager):"
        echo "  curl -sSL https://git.io/g-install | sh -s"
        echo "  или скачайте напрямую: https://golang.org/dl/"
        exit 1
    fi
    
    echo "✅ Go: $(go version)"
    
    # Проверяем, установлен ли g
    if command -v g >/dev/null 2>&1; then
        echo "✅ g (Golang Version Manager): $(g version)"
    fi
}

# Функция для проверки PATH
check_path() {
    echo "🔍 Проверка PATH..."
    
    GOPATH=$(go env GOPATH)
    if [ -z "$GOPATH" ]; then
        echo "⚠️  GOPATH не установлен"
        exit 1
    fi
    
    echo "📁 GOPATH: $GOPATH"
    
    # Проверяем, есть ли GOPATH/bin в PATH
    if [[ ":$PATH:" == *":$GOPATH/bin:"* ]]; then
        echo "✅ PATH настроен корректно"
    else
        echo "⚠️  GOPATH/bin не найден в PATH"
        echo "💡 Выполните: make path"
        exit 1
    fi
}

# Функция для проверки Docker (опционально)
check_docker() {
    if command -v docker >/dev/null 2>&1; then
        echo "✅ Docker: $(docker --version)"
    else
        echo "⚠️  Docker не установлен (опционально для разработки)"
    fi
}

# Функция для предложения установки расширений
suggest_extensions() {
    echo "🔧 Расширения VS Code:"
    echo "  📝 Отредактируйте script/ext.sh по своему усмотрению"
    echo "  🚀 Выполните: task ext"
}

# Основная логика
main() {
    echo "🔍 Проверка зависимостей..."
    
    # Проверяем основные команды
    check_go
    check_path
    check_docker
    
    # Устанавливаем системные зависимости
    case "$PLATFORM" in
        linux)
            install_linux_deps
            ;;
        macos)
            install_macos_deps
            ;;
        windows)
            install_windows_deps
            ;;
    esac
    
    # Настраиваем конфигурацию
    setup_config
    
    # Устанавливаем Go инструменты
    install_go_tools
    
    # Синхронизируем Go модули
    echo "📦 Синхронизация Go модулей..."
    task mod
    
    # Предлагаем установить расширения VS Code
    suggest_extensions
    
    echo ""
    echo "🎉 Инициализация завершена!"
    echo ""
    echo "📋 Основные команды:"
    echo "  task - список команд"
    echo "  task <name> --summary - описание команды <name>"
}

# Запускаем основную функцию
main "$@" 