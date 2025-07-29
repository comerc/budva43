#!/bin/sh
set -e  # выход при любой ошибке
set -o pipefail  # выход при ошибке в любой команде пайплайна

echo "🔍 Поиск удаленных файлов в истории Git..."
echo "💡 Используем git-filter-repo - современный и быстрый инструмент"

# Создаем временные файлы
TEMP_DIR=$(mktemp -d)
ALL_FILES="$TEMP_DIR/all_files.txt"
DELETED_FILES="$TEMP_DIR/deleted_files.txt"

# Получаем все файлы из истории только текущей ветки
git log --name-only --pretty=format: | sort -u | grep -v "^$" > "$ALL_FILES"

echo "🔍 Всего файлов в истории текущей ветки: $(wc -l < "$ALL_FILES")"

# Находим удаленные файлы (физически удаленные или в .gitignore)
> "$DELETED_FILES"
while IFS= read -r file; do
    # Файл удален физически
    if [ ! -f "$file" ] && [ ! -d "$file" ]; then
        echo "$file" >> "$DELETED_FILES"
    # Файл существует, но игнорируется Git
    elif git check-ignore "$file" >/dev/null 2>&1; then
        echo "$file" >> "$DELETED_FILES"
    fi
done < "$ALL_FILES"

echo ""
echo "📋 Найденные удаленные файлы:"
if [ -s "$DELETED_FILES" ]; then
    cat "$DELETED_FILES"
    DELETED_COUNT=$(wc -l < "$DELETED_FILES")
    echo ""
    echo "📊 Всего: $DELETED_COUNT файлов"
else
    echo "✅ Удаленных файлов не найдено!"
    rm -rf "$TEMP_DIR"
    exit 0
fi

echo ""
read -p "🤔 Удалить эти файлы из истории текущей ветки? (y/N): " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Проверяем git-filter-repo
    if ! command -v git-filter-repo >/dev/null 2>&1; then
        echo "❌ git-filter-repo не найден!"
        rm -rf "$TEMP_DIR"
        exit 1
    fi
    
    echo "⚠️  ВНИМАНИЕ: Это изменит историю текущей ветки!"
    read -p "🔄 Продолжить? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🧹 Удаление файлов из истории с помощью git-filter-repo..."
        
        while IFS= read -r file; do
            echo "🗑️  Удаляем: $file"
            git filter-repo --path "$file" --invert-paths --refs HEAD --force
        done < "$DELETED_FILES"
        
        echo "✅ Очистка завершена!"
        echo "⚠️  Выполните: git push origin --force"
    fi
fi

rm -rf "$TEMP_DIR"