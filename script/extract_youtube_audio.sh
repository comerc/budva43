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
TEMP_DIR="/tmp/_extract_youtube_audio_$$"
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

echo "Скачиваем аудио с YouTube..."
yt-dlp -f "bestaudio" -o "source.%(ext)s" "$HASH"

# Определяем расширение скачанного файла
SOURCE_FILE=$(ls source.*)
EXT="${SOURCE_FILE#*.}"
TRIMMED_FILE="trimmed.$EXT"

echo "Отрезаем начало ($SKIP_TIME)..."
ffmpeg -ss "$SKIP_TIME" -i "$SOURCE_FILE" -c copy "$TRIMMED_FILE"

echo "Конвертируем в битрейт 48k..."
ffmpeg -i "$TRIMMED_FILE" -c:a aac -b:a 48k final.m4a

# Получаем длительность оставшегося аудио
DURATION=$(ffprobe -v quiet -show_entries format=duration -of csv=p=0 final.m4a)
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
        ffmpeg -ss $START_TIME -t $CHUNK_DURATION -i final.m4a -c copy "$OUTPUT_FILE"
    else
        # Последний кусок (остаток)
        REMAINING=$((DURATION_INT - START_TIME))
        echo "Создаём последний кусок $CHUNK_NUM ($REMAINING секунд)..."
        ffmpeg -ss $START_TIME -i final.m4a -c copy "$OUTPUT_FILE"
    fi
    
    START_TIME=$((START_TIME + CHUNK_DURATION))
    CHUNK_NUM=$((CHUNK_NUM + 1))
done

echo "Перемещаем файлы в папку _extract_youtube_audio..."
# Возвращаемся в исходную папку
cd "$ORIGINAL_DIR"

# Создаём папку _extract_youtube_audio если её нет
mkdir -p _extract_youtube_audio

# Удаляем старые файлы output*.m4a если они есть
rm -f _extract_youtube_audio/output*.m4a

# Перемещаем новые файлы
mv "$TEMP_DIR"/output*.m4a _extract_youtube_audio/

echo "Готово! Создано $((CHUNK_NUM - 1)) файлов в папке _extract_youtube_audio/"
ls -la _extract_youtube_audio/output*.m4a