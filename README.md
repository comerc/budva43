# budva43

[![Go Version](https://img.shields.io/github/go-mod/go-version/comerc/budva43)](https://go.dev/doc/install)
[![Go Report Card](https://goreportcard.com/badge/github.com/comerc/budva43)](https://goreportcard.com/report/github.com/comerc/budva43)
[![codecov](https://codecov.io/gh/comerc/budva43/graph/badge.svg?token=JGTZM00AXV)](https://codecov.io/gh/comerc/budva43)
[![Last Commit](https://img.shields.io/github/last-commit/comerc/budva43)](https://github.com/comerc/budva43/commits/main/)
[![Project Status](https://img.shields.io/github/release/comerc/budva43.svg)](https://github.com/comerc/budva43/releases/latest)

Telegram-Forwarder (UNIX-way) - forwards (or copies) messages from monitored channels and groups to target ones according to specified rules to obtain thematic digests.

## How to clone project with submodules

```bash
$ git clone https://github.com/comerc/budva43.git
$ git submodule init
$ git submodule update
```

## How to Dev Start

Direct install TDLib on host machine for best dev experience or use DevContainer (some restrictions) for build on Ubuntu.

### With direct TDLib

Install by [instruction](https://github.com/zelenin/go-tdlib/blob/master/README.md) with this options:

- Install built TDLib to /usr/local instead of placing the files to td/tdlib.
- Choose which compiler you want to use to build TDLib: clang (recommended)

### With DevContainer

only first time:
```bash
$ docker-compose build
```
...then "Reopen in Container"

### Install Mockery V2

```bash
$ go install github.com/vektra/mockery/v2@v2.53.3
```

## Applied Technologies

- [x] [Dev Containers](https://code.visualstudio.com/docs/devcontainers/containers)
- [x] [testcontainers](https://testcontainers.com/guides/getting-started-with-testcontainers-for-go/)
- [x] [mockery](https://github.com/vektra/mockery)
- [ ] [easyjson](https://github.com/mailru/easyjson)
- [x] Docker Compose
- [x] [docker multi-stage build](https://docs.docker.com/build/building/multi-stage/)
- [x] zsh + [ohmyzsh](https://ohmyz.sh/)
- [x] golangci-lint + [revive](https://revive.run/)
- [x] Makefile
- [x] [editorconfig](https://editorconfig.org/)
- [x] [git submodule](https://git-scm.com/book/en/v2/Git-Tools-Submodules)
- [ ] [goreleaser](https://goreleaser.com/)
- [x] [todo-highlight](https://marketplace.visualstudio.com/items?itemName=wayou.vscode-todo-highlight)
- [x] [git-graph](https://marketplace.visualstudio.com/items?itemName=mhutchie.git-graph)
- [ ] Redis
- [x] [go-clean-architecture](https://github.com/comerc/go-clean-architecture)
- [x] SOLID
- [x] DRY
- [x] KISS
- [x] YAGNI
- [x] [Dependency Injection](https://habr.com/ru/companies/vivid_money/articles/531822/)
- [x] [design-patterns](https://refactoring.guru/ru/design-patterns/go)
- [x] Notations Start & Run
- [x] [Graceful Shutdown](https://habr.com/ru/articles/771626/)
- [ ] [Front Controller (microservice)](https://en.wikipedia.org/wiki/Front_controller)
- [ ] CQRS
- [x] [samber/lo](https://github.com/samber/lo)
- [ ] [uptrace/bun](https://github.com/uptrace/bun)
- [x] [gqlgen](https://gqlgen.com/)
- [ ] Grafana + Prometheus
- [ ] ClickHouse
- [ ] [fiber](https://gofiber.io/)
- [x] [zelenin/go-tdlib](https://github.com/zelenin/go-tdlib)
- [ ] [Code Style](https://github.com/quickwit-oss/quickwit/blob/206ebf791af78f11c562835a449df449b3a17e81/CODE_STYLE.md
)
- [ ] samber/mo
- [ ] samber/do
- [ ] samber/oops
- [ ] samber/slog-*
- [x] testing: Time & Concurrency https://danp.net/posts/synctest-experiment/
- [ ] spf13/cobra
- [x] go.mod replace
- [x] [comerc/spylog](https://github.com/comerc/spylog)
- [x] go test -race WARNING LC_DYSYMTAB https://github.com/golang/go/issues/61229
- [x] voidint/g
- [x] go mod vendor
- [x] [vmihailenco/msgpack](https://github.com/vmihailenco/msgpack)
- [ ] temporal
- [ ] anycable
- [x] [gocritic](https://habr.com/ru/articles/414739/)
- [ ] [OpenTelemetry](https://pkg.go.dev/go.opentelemetry.io/otel#section-readme)
- [ ] [gostackparse](https://github.com/DataDog/gostackparse)
- [x] grafana-loki
- [x] [pplog](https://github.com/michurin/human-readable-json-logging)
- [ ] pprof
- [x] time.AfterFunc() & context.AfterFunc()
- [x] init() - dependency chain: config -> log -> spylog
- [x] closures (currying and partial application)
- [x] tail recursion
- [x] errors.Is & errors.As
- [x] generics
- [x] PUB/SUB
- [x] table (call-)tests
- [x] integration tests
- [x] separate config for e2e tests
- [x] dynamic config engine.yml
- [x] structured logs and errors
- [x] termAutomator
- [ ] codecov.io (as uber/atomic)
- [x] snapshot tests
- [ ] [LocalAI](https://github.com/mudler/LocalAI)
- [ ] [go-prompt](https://github.com/c-bata/go-prompt)
- [ ] Gemma for NLP?
- [ ] uber-go/automaxprocs
- [ ] uber-go/goleak
- [ ] magefile/mage
- [ ] oklog/ulid
- [x] [Go to Implementation](https://github.com/comerc/golang-go-to-impl)
- [x] [jq](https://jqlang.org/)
- [x] [task](https://taskfile.dev/)
- [x] [unused-interface-methods](https://github.com/unused-interface-methods/unused-interface-methods)
- [x] BDD
- [x] [godog](github.com/cucumber/godog)
- [x] grpc
- [ ] fuzz-test?



## .env

[Register an application](https://my.telegram.org/apps) to obtain an api_id and api_hash

```ini
# ./config/.private/.env

BUDVA43__TELEGRAM__API_ID=1234567
BUDVA43__TELEGRAM__API_HASH=XXXXXXX
BUDVA43__TELEGRAM__PHONE_NUMBER=+78901234567
```

## Config example

```bash
./config/app.yml
./config/engine.yml
```

## First start for Telegram auth

```bash
$ make run
```

<!--

## Old variants for Telegram auth (draft)

via web:

http://localhost:7007


or via docker:

```
$ make
$ make up
$ docker attach telegram-forwarder
```

but then we have problem with permissions (may be need docker rootless mode?):

```
$ sudo chmod -R 777 ./tdata
```

## Get chat list with limit (optional)

http://localhost:7007?limit=10

-->

## Examples for go-tdlib

```go
// How to add InlineKeyboardButton

	row := make([]*client.InlineKeyboardButton, 0)
	row = append(row, &client.InlineKeyboardButton{
		Text: "123",
		Type: &client.InlineKeyboardButtonTypeUrl{
			Url: "https://google.com",
		},
	})
	rows := make([][]*client.InlineKeyboardButton, 0)
	rows = append(rows, row)
	_, err := tdlibClient.SendMessage(&client.SendMessageRequest{
		ChatId: dstChatId,
		InputMessageContent: &client.InputMessageText{
			Text:                  formattedText,
			DisableWebPagePreview: true,
			ClearDraft:            true,
		},
		ReplyMarkup: &client.ReplyMarkupInlineKeyboard{
			Rows: rows,
		},
	})

```

## Inspired by

- [marperia/fwdbot](https://github.com/marperia/fwdbot)
- [wcsiu/telegram-client-demo](https://github.com/wcsiu/telegram-client-demo) + [article](https://wcsiu.github.io/2020/12/26/create-a-telegram-client-in-go-with-docker.html)
- [Создание и развертывание ретранслятора Telegram каналов, используя Python и Heroku](https://vc.ru/dev/158757-sozdanie-i-razvertyvanie-retranslyatora-telegram-kanalov-ispolzuya-python-i-heroku)

## Filters Mode for Forward...

```
Exclude #COIN
Include #TSLA

case #COIN
Check +
Other -
To -

case #TSLA
Check -
Other -
To +

case #ARK
Check -
Other +
To -
```

## Test-plan for Config...

- Text
  - [x] Forward.SendCopy (or forward)
  - [x] and edit sync for double copy
  - [x] Forward.CopyOnce (edit sync)
  - [x] Forward.Indelible (delete sync)
  - [x] Filters Mode (see above)
  - [x] Forward.IncludeSubmatch
  - [x] ReplaceMyselfLinks + DeleteExternal
  - [x] ReplaceFragments (and not equal len)
  - [x] Sources.Link + Title
  - [x] Sources.Sign
  - [ ] AutoAnswers
- MediaAlbum
  - [x] Forward.SendCopy (or forward)
  - [x] Forward.CopyOnce (edit sync)
  - [x] Forward.Indelible (delete sync)

## Dependencies

- [task](https://taskfile.dev/)
- [pplog](https://github.com/michurin/human-readable-json-logging)
- [tail](https://github.com/uutils/coreutils) (cross platform)
- [gojq](https://github.com/itchyny/gojq)
- [jqp](https://github.com/noahgorstein/jqp)
- [golangci-lint](https://github.com/golangci/golangci-lint)
- [godog](github.com/cucumber/godog)
- [protobuf](https://github.com/protocolbuffers/protobuf-go)

```bash
# golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# godog
go install github.com/cucumber/godog/cmd/godog@latest

# protobuf
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## How to check grpc

```bash
brew install grpcurl
grpcurl -plaintext localhost:50051 list
```
