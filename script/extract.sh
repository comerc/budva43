#!/bin/sh
set -e  # выход при любой ошибке
set -o pipefail  # выход при ошибке в любой команде пайплайна

# Проверяем количество аргументов
if [ $# -ne 2 ]; then
    echo "Использование: $0 HASH XX:XX"
    echo "HASH - ключ видео на YouTube"
    echo "XX:XX - сколько минут отсекать от начала ролика"
    exit 1
fi

HASH=$1
SKIP_TIME=$2

# Сохраняем текущую папку
ORIGINAL_DIR=$(pwd)

# Создаём временную папку
TEMP_DIR="/tmp/_extract_$$"
mkdir -p "$TEMP_DIR"

# Функция очистки при выходе
cleanup() {
    echo "Очистка временных файлов..."
    cd "$ORIGINAL_DIR" > /dev/null 2>&1 || true
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# Переходим во временную папку
cd "$TEMP_DIR"

echo "Скачиваем видео с YouTube..."
yt-dlp -x --audio-format m4a -o download.m4a "$HASH"

echo "Конвертируем в битрейт 48k..."
ffmpeg -i download.m4a -c:a aac -b:a 48k input.m4a

echo "Отрезаем начало ($SKIP_TIME)..."
ffmpeg -ss "$SKIP_TIME" -i input.m4a -c copy output.m4a

# Получаем длительность оставшегося аудио
DURATION=$(ffprobe -v quiet -show_entries format=duration -of csv=p=0 output.m4a)
DURATION_INT=$(echo "$DURATION" | cut -d. -f1)

# Проверяем, что получили корректную длительность
if [ -z "$DURATION_INT" ] || [ "$DURATION_INT" -le 0 ]; then
    echo "Ошибка: не удалось определить длительность аудио"
    exit 1
fi

echo "Длительность после обрезки: $DURATION_INT секунд"

# Нарезаем кусками по 20 минут (1200 секунд)
CHUNK_DURATION=1200
CHUNK_NUM=1
START_TIME=0

while [ $START_TIME -lt $DURATION_INT ]; do
    OUTPUT_FILE="output${CHUNK_NUM}.m4a"
    
    if [ $((START_TIME + CHUNK_DURATION)) -lt $DURATION_INT ]; then
        # Полный кусок 20 минут
        echo "Создаём кусок $CHUNK_NUM (20 минут)..."
        ffmpeg -ss $START_TIME -t $CHUNK_DURATION -i output.m4a -c copy "$OUTPUT_FILE"
    else
        # Последний кусок (остаток)
        REMAINING=$((DURATION_INT - START_TIME))
        echo "Создаём последний кусок $CHUNK_NUM ($REMAINING секунд)..."
        ffmpeg -ss $START_TIME -i output.m4a -c copy "$OUTPUT_FILE"
    fi
    
    START_TIME=$((START_TIME + CHUNK_DURATION))
    CHUNK_NUM=$((CHUNK_NUM + 1))
done

# Удаляем промежуточный файл output.m4a (он больше не нужен)
rm -f output.m4a

echo "Перемещаем файлы в папку _extract..."
# Возвращаемся в исходную папку
cd "$ORIGINAL_DIR"

# Создаём папку _extract если её нет
mkdir -p _extract

# Удаляем старые файлы output*.m4a если они есть
rm -f _extract/output*.m4a

# Перемещаем новые файлы
mv "$TEMP_DIR"/output*.m4a _extract/

echo "Готово! Создано $((CHUNK_NUM - 1)) файлов в папке _extract/"
ls -la _extract/output*.m4a