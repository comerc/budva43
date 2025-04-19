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

all:
	lint
	test
	clean
	build

