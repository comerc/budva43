# Пути к библиотеке tdlib
TD_INCLUDE_PATH = /td/tdlib/include
TD_LIB_PATH = /td/tdlib/lib

# Общие переменные окружения для сборки и линтинга - инициализируются в docker-compose.yml
COMMON_ENV = CGO_CFLAGS=-I$(TD_INCLUDE_PATH) CGO_LDFLAGS="-Wl,-rpath,$(TD_LIB_PATH) -L$(TD_LIB_PATH) -ltdjson"

lint:
	golangci-lint run

test:
	go test -v ./...

clean:
	rm -f bin/app

build:
	go build -o bin/app main.go

run:
	go run main.go

all:
	lint
	test
	clean
	build

