#!/bin/bash
set -e  # выход при любой ошибке
set -o pipefail  # выход при ошибке в любой команде пайплайна

echo "🔧 Установка расширений VS Code..."

# Проверяем наличие VS Code
if ! command -v code >/dev/null 2>&1; then
    echo "⚠️  VS Code не найден в PATH"
    echo "💡 Установите VS Code: https://code.visualstudio.com/"
    exit 1
fi

# Список расширений с комментариями
EXTENSIONS=(
    "golang.go"                           # Go разработка
    "ethan-reesor.exp-vscode-go"          # Дополнительные возможности для Go
    "comerc.go-table-test-navigator"      # Навигация по табличным тестам
    "comerc.golang-go-to-impl"            # Переход к реализации интерфейсов
    "graphql.vscode-graphql"              # GraphQL поддержка
    "graphql.vscode-graphql-syntax"       # GraphQL синтаксис
    "alexkrechik.cucumberautocomplete"    # BDD тестирование
    "anysphere.pyright"                   # Python поддержка
    "ms-azuretools.vscode-docker"         # Docker поддержка
    "github.vscode-github-actions"        # GitHub Actions
    "eamodio.gitlens"                     # Расширенная работа с Git
    "wayou.vscode-todo-highlight"         # Подсветка TODO комментариев
    "zxh404.vscode-proto3"                # Protocol Buffers поддержка
    "formulahendry.auto-rename-tag"       # Автопереименование тегов
    "adrianwilczynski.toggle-hidden"      # Переключение скрытых файлов
    "jellydn.toggle-excluded-files"       # Переключение исключенных файлов
    "wakatime.vscode-wakatime"            # Статистика времени разработки
)

echo "📦 Устанавливаем ${#EXTENSIONS[@]} расширений..."

for extension in "${EXTENSIONS[@]}"; do
    # Извлекаем комментарий (все после #)
    comment=$(echo "$extension" | sed 's/.*# //')
    # Извлекаем ID расширения (все до #)
    ext_id=$(echo "$extension" | sed 's/#.*//')
    
    echo "  📥 $ext_id - $comment"
    code --install-extension "$ext_id" --force >/dev/null 2>&1 || true
done

echo "✅ Расширения VS Code установлены" 