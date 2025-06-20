all: lint check build

lint:
	GOEXPERIMENT=synctest golangci-lint run

check:
	GOEXPERIMENT=synctest go test -short -failfast -race -count=1 ./...

build:
	rm -f bin/app && go build -o bin/app main.go

run:
	@BUDVA43__GENERAL__ENGINE_CONFIG_FILE=engine.e2e.yml \
	go run main.go

kill-port:
	lsof -ti:7070 | xargs kill -9

test-auth-telegram-state:
	curl http://localhost:7070/api/auth/telegram/state

test-auth-telegram-submit-phone:
	@read -s -p "Введите номер телефона (скрытый ввод): " PHONE && echo && \
	curl --header "Content-Type: application/json" \
		--request POST \
		--data '{"phone":"'$$PHONE'"}' \
		http://localhost:7070/api/auth/telegram/phone

test-auth-telegram-submit-code:
	@read -s -p "Введите код (скрытый ввод): " CODE && echo && \
	curl --header "Content-Type: application/json" \
		--request POST \
		--data '{"code":"'$$CODE'"}' \
		http://localhost:7070/api/auth/telegram/code

test-auth-telegram-submit-password:
	@read -s -p "Введите пароль (скрытый ввод): " PASSWORD && echo && \
	curl --header "Content-Type: application/json" \
		--request POST \
		--data '{"password":"'$$PASSWORD'"}' \
		http://localhost:7070/api/auth/telegram/password
