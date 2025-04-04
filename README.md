# budva43

Telegram-Forwarder (UNIX-way)

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
		fmt.Print("GetMessage src ", err)
	} else {
		messageLink, err := tdlibClient.GetMessageLink(&client.GetMessageLinkRequest{
			ChatId:     src.ChatId,
			MessageId:  src.Id,
			ForAlbum:   src.MediaAlbumId != 0,
			ForComment: false,
		})
		if err != nil {
			fmt.Print("GetMessageLink ", err)
		} else {
			fmt.Print(messageLink.Link)
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
