#!/bin/bash
set -e  # выход при любой ошибке
set -o pipefail  # выход при ошибке в любой команде пайплайна

mkdir -p .coverage

# не отключать COVERAGE_EXCLUDE на этом этапе
GOEXPERIMENT=synctest go test -covermode=atomic -coverprofile=.coverage/.out -coverpkg=./... ./...

COVERAGE_EXCLUDE="(/mocks/|_easyjson\.go|/graph/|/pb/)"
grep -vE "$COVERAGE_EXCLUDE" .coverage/.out > .coverage/.txt
rm .coverage/.out
