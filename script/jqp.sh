#!/bin/sh
set -e  # выход при любой ошибке
set -o pipefail  # выход при ошибке в любой команде пайплайна

jqp -f .data/$SUBPROJECT/log/app.log