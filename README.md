# budva43

Telegram-Forwarder (UNIX-way)

[PLAN](./docs/PLAN.md)

## How to clone project with submodules

```bash
git clone https://github.com/comerc/budva43.git
git submodule init
git submodule update
```

## How to Dev Start 

Use DevContainer (some restrictions) for build on Ubuntu or direct install TDLib on host machine for best dev experience.

### With DevContainer

only first time:
```bash
docker-compose build
```
...then "Reopen in Container"

### With direct TDLib 

Install by [instruction](https://github.com/zelenin/go-tdlib/blob/master/README.md) with this options:

- Install built TDLib to /usr/local instead of placing the files to td/tdlib.
- Choose which compiler you want to use to build TDLib: clang (recommended)

## Applied Technologies

- [x] [Dev Containers](https://code.visualstudio.com/docs/devcontainers/containers)
- [ ] [testcontainers](https://testcontainers.com/guides/getting-started-with-testcontainers-for-go/)
- [x] [mockery](https://github.com/vektra/mockery)
- [ ] [easyjson](https://github.com/mailru/easyjson)
- [x] Docker Compose
- [x] [multi-stage build](https://docs.docker.com/build/building/multi-stage/)
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
- [x] DRY (без фанатизма) 
- [x] KISS
- [x] [Dependency Injection](https://habr.com/ru/companies/vivid_money/articles/531822/)
- [x] [design-patterns](https://refactoring.guru/ru/design-patterns/go)
- [x] Нотации start и run
- [x] [Graceful Shutdown](https://habr.com/ru/articles/771626/)
- [ ] [Front Controller (microservice)](https://en.wikipedia.org/wiki/Front_controller)
- [ ] CQRS
- [ ] [samber/lo](https://github.com/samber/lo)
- [ ] [uptrace/bun](https://github.com/uptrace/bun)
- [ ] [gqlgen](https://gqlgen.com/)
- [ ] Grafana + Prometheus
- [ ] ClickHouse
- [ ] [fiber](https://gofiber.io/)
- [x] [zelenin/go-tdlib](https://github.com/zelenin/go-tdlib)
- [ ] [Code Style](https://github.com/quickwit-oss/quickwit/blob/206ebf791af78f11c562835a449df449b3a17e81/CODE_STYLE.md
)
- [ ] samber/mo
- [ ] samber/lo
- [ ] samber/do
- [ ] samber/oops
- [ ] samber/slog-*
- [ ] тестирование Time & Concurrency
- [ ] spf13/cobra
- [x] go.mod replace
- [ ] comerc/slog-module
- [x] go test -race WARNING LC_DYSYMTAB https://github.com/golang/go/issues/61229
- [x] voidint/g

## .env

[Register an application](https://my.telegram.org/apps) to obtain an api_id and api_hash

```
BUDVA32_API_ID=1234567
BUDVA32_API_HASH=XXXXXXXX
BUDVA32_PHONENUMBER=78901234567
BUDVA32_PORT=4004
```

## First start for Telegram auth via web

http://localhost:4004

<!-- ## Old variants for Telegram auth (draft)

from console:

```
$ go run .
```

or via docker:

```
$ make
$ make up
$ docker attach telegram-forwarder
```

but then we have problem with permissions (may be need docker rootless mode?):

```
$ sudo chmod -R 777 ./tdata
``` -->

## .config.yml example

```yml
# escape markdown '\*\_\{\}\[\]\(\)\#\+\-\.\!'
ReplaceMyselfLinks: # for destinations
	-2222:
    DeleteExternal: true
ReplaceFragments: # for destinations
  -2222:
    "aaaa": "bbbb" # must be equal length
Sources:
  -1234:
    Sign:
      Title: '*\#Source*' # for SendCopy (with markdown)
      For: [-8888]
    Link:
			Title: "*Source*" # for SendCopy (with markdown)
			For: [-4321]
Reports:
  Template: "За *24 часа* отобрал: *%d* из *%d* 😎\n\\#ForwarderStats" # (with markdown)
  For: [
      -2222,
      -4321,
      -8888,
    ]
Forwards:
	"id1":
		From: -1111
		To: [-2222]
	"id2":
		From: -1234
		To: [-4321, -8888]
		SendCopy: true
		CopyOnce: true # wo edit-sync
		Indelible: true # wo delete-sync
		Exclude: 'Крамер|#УТРЕННИЙ_ОБЗОР'
		Include: '#ARK|#Идеи_покупок|#ОТЧЕТЫ'
		IncludeSubmatch:
			- Regexp: '(^|[^A-Z])\$([A-Z]+)'
				Group: 2
				Match: ['F', 'GM', 'TSLA']
		Other: -4444 # after include (copy only)
		Check: -7777 # after exclude (forward only)
```

## Get chat list with limit (optional)

http://localhost:4004?limit=10

## Examples for go-tdlib

```go
func getMessageLink(srcChatId, srcMessageId int) {
	src, err := tdlibClient.GetMessage(&client.GetMessageRequest{
		ChatId:    int64(srcChatId),
		MessageId: int64(srcMessageId),
	})
	if err != nil {
		slog.Error("GetMessage src", "err", err)
	} else {
		messageLink, err := tdlibClient.GetMessageLink(&client.GetMessageLinkRequest{
			ChatId:     src.ChatId,
			MessageId:  src.Id,
			ForAlbum:   src.MediaAlbumId != 0,
			ForComment: false,
		})
		if err != nil {
			slog.Error("GetMessageLink", "err", err)
		} else {
			slog.Info("GetMessageLink", "link", messageLink.Link)
		}
	}
}

// How to use update?

	for update := range listener.Updates {
		if update.GetClass() == client.ClassUpdate {
			if updateNewMessage, ok := update.(*client.UpdateNewMessage); ok {
				//
			}
		}
	}

// etc
// https://github.com/zelenin/go-tdlib/blob/ec36320d03ff5c891bb45be1c14317c195eeadb9/client/type.go#L1028-L1108

// How to use markdown?

	formattedText, err := tdlibClient.ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: "*bold* _italic_ `code`",
		ParseMode: &client.TextParseModeMarkdown{
			Version: 2,
		},
	})
	if err != nil {
		log.Print(err)
	} else {
		log.Printf("%#v", formattedText)
	}

// How to add InlineKeyboardButton

	row := make([]*client.InlineKeyboardButton, 0)
	row = append(row, &client.InlineKeyboardButton{
		Text: "1234",
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
