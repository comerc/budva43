#!/bin/sh
set -e  # выход при любой ошибке
set -o pipefail  # выход при ошибке в любой команде пайплайна

# Если в командной строке есть аргументы, используем их все как фильтр.
# Если аргументов нет, фильтром по умолчанию будет 'true' (показывать все).
if [ $# -gt 0 ]; then
  FILTER="$*"
else
  FILTER='true'
fi

# Запускаем основной конвейер с фильтром из аргументов.
tail $TAIL_ARGS -F .data/log/app.log | gojq -R -r "if (fromjson | ${FILTER}) then . else empty end" | PROJECT_ROOT=$(pwd)/ pplog
