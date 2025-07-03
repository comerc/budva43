#!/bin/sh

# Если в командной строке есть аргументы, используем их все как фильтр.
# Если аргументов нет, фильтром по умолчанию будет 'true' (показывать все).
if [ $# -gt 0 ]; then
  filter="$*"
else
  filter='true'
fi

# Запускаем основной конвейер с фильтром из аргументов.
gtail -F .data/log/app.log | jq -R -r --unbuffered "if (fromjson | ${filter}) then . else empty end" | PROJECT_ROOT=$(pwd)/ pplog
