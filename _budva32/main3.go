package main

import (
	"crypto/subtle"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/zelenin/go-tdlib/client"
)

type FiltersMode string

const (
	FiltersOK    FiltersMode = "ok"
	FiltersCheck FiltersMode = "check"
	FiltersOther FiltersMode = "other"
)

// OK: перенесено - service/engine/service.go (mapFiltersMode)
func checkFilters(formattedText *client.FormattedText, forward config.Forward) FiltersMode {
	if formattedText.Text == "" {
		hasInclude := false
		if forward.Include != "" {
			hasInclude = true
		}
		for _, includeSubmatch := range forward.IncludeSubmatch {
			if includeSubmatch.Regexp != "" {
				hasInclude = true
				break
			}
		}
		if hasInclude {
			return FiltersOther
		}
	} else {
		if forward.Exclude != "" {
			re := regexp.MustCompile("(?i)" + forward.Exclude)
			if re.FindString(formattedText.Text) != "" {
				return FiltersCheck
			}
		}
		hasInclude := false
		if forward.Include != "" {
			hasInclude = true
			re := regexp.MustCompile("(?i)" + forward.Include)
			if re.FindString(formattedText.Text) != "" {
				return FiltersOK
			}
		}
		for _, includeSubmatch := range forward.IncludeSubmatch {
			if includeSubmatch.Regexp != "" {
				hasInclude = true
				re := regexp.MustCompile("(?i)" + includeSubmatch.Regexp)
				matches := re.FindAllStringSubmatch(formattedText.Text, -1)
				for _, match := range matches {
					s := match[includeSubmatch.Group]
					if contains(includeSubmatch.Match, s) {
						return FiltersOK
					}
				}
			}
		}
		if hasInclude {
			return FiltersOther
		}
	}
	return FiltersOK
}

// НЕТ: не перенесено, предлагаю - transport/web/middleware.go (withBasicAuth)
func withBasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(user), []byte(os.Getenv("BUDVA32_USER"))) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(os.Getenv("BUDVA32_PASS"))) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Please enter your username and password"`)
			http.Error(w, "You are unauthorized to access the application.\n", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}

// НЕТ: не перенесено, предлагаю - transport/web/middleware.go (withAuthentiation)
func withAuthentiation(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if outputCh != nil {
			if r.Method == "POST" {
				r.ParseForm()
				if len(r.PostForm["input"]) == 1 {
					input := r.PostForm["input"][0]
					inputCh <- input
				}
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			output := <-outputCh
			if output != "Ready!" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				io.WriteString(w, fmt.Sprintf(`<html><head><title>%s</title></head><body><form method="post">%s<input autocomplete="off" name="input" /><input type="submit" /></form></body></html>`, projectName, output))
				return
			}
			outputCh = nil
		}
		handler(w, r)
	}
}

// NO: не перенесено, предлагаю - util/network.go (getIP)
func getIP() string {
	interfaces, _ := net.Interfaces()
	for _, i := range interfaces {
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			return ip.String()
		}
	}
	return ""
}

// НЕТ: не перенесено, предлагаю - transport/web/transport.go (getFaviconHandler)
func getFaviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/favicon.ico")
}

// НЕТ: не перенесено, предлагаю - transport/web/transport.go (getAnswerHandler)
func getAnswerHandler(w http.ResponseWriter, r *http.Request) {
	// использует коды ошибок HTTP для статусов: error, ok, wait
	// 200 OK
	// 204 No Content
	// 205 Reset Content
	// 500 Internal Server Error
	// TODO: накапливать статистику по параметру step, чтобы подкрутить паузу в shunt
	q := r.URL.Query()
	log.Printf("getAnswerHandler > %#v", q)
	var isOnlyCheck bool
	if len(q["only_check"]) == 1 {
		isOnlyCheck = q["only_check"][0] == "1"
	}
	var dstChatId int64
	if len(q["chat_id"]) == 1 {
		dstChatId = int64(convertToInt(q["chat_id"][0]))
	}
	var newMessageId int64
	if len(q["message_id"]) == 1 {
		newMessageId = int64(convertToInt(q["message_id"][0]))
	}
	if dstChatId == 0 || newMessageId == 0 {
		err := fmt.Errorf("invalid input parameters")
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpMessageId := getTmpMessageId(dstChatId, newMessageId)
	if tmpMessageId == 0 {
		log.Print("http.StatusNoContent")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	fromChatMessageId := getAnswerMessageId(dstChatId, tmpMessageId)
	if fromChatMessageId == "" {
		log.Print("http.StatusResetContent #1")
		w.WriteHeader(http.StatusResetContent)
		return
	}
	a := strings.Split(fromChatMessageId, ":")
	srcChatId := int64(convertToInt(a[0]))
	srcMessageId := int64(convertToInt(a[1]))
	message, err := tdlibClient.GetMessage(&client.GetMessageRequest{
		ChatId:    srcChatId,
		MessageId: srcMessageId,
	})
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, ok := getReplyMarkupData(message)
	if !ok {
		log.Print("http.StatusResetContent #2")
		w.WriteHeader(http.StatusResetContent)
		return
	}
	if !isOnlyCheck {
		answer, err := tdlibClient.GetCallbackQueryAnswer(&client.GetCallbackQueryAnswerRequest{
			ChatId:    srcChatId,
			MessageId: srcMessageId,
			Payload:   &client.CallbackQueryPayloadData{Data: data},
		})
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, answer.Text)
	}
}

// НЕТ: не перенесено, предлагаю - transport/web/transport.go (getPingHandler)
func getPingHandler(w http.ResponseWriter, r *http.Request) {
	ret, err := time.Now().UTC().MarshalJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, fmt.Sprintf("{now:%s}", string(ret)))
}

// НЕТ: не перенесено, предлагаю - transport/web/transport.go (getChatsHandler)
func getChatsHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var limit = 1000
	if len(q["limit"]) == 1 {
		limit = convertToInt(q["limit"][0])
	}
	allChats, err := getChatList(tdlibClient, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	retMap := make(map[string]interface{})
	retMap["total"] = len(allChats)
	var chatList []string
	for _, chat := range allChats {
		chatList = append(chatList, fmt.Sprintf("%d=%s", chat.Id, chat.Title))
	}
	retMap["chatList"] = chatList
	ret, err := json.Marshal(retMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(ret))
}

// see https://stackoverflow.com/questions/37782348/how-to-use-getchats-in-tdlib
// НЕТ: не перенесено, предлагаю - service/telegram/service.go (getChatList)
func getChatList(tdlibClient *client.Client, limit int) ([]*client.Chat, error) {
	var (
		allChats     []*client.Chat
		offsetOrder  = int64(math.MaxInt64)
		offsetChatId = int64(0)
	)
	for len(allChats) < limit {
		if len(allChats) > 0 {
			lastChat := allChats[len(allChats)-1]
			for i := 0; i < len(lastChat.Positions); i++ {
				if lastChat.Positions[i].List.ChatListType() == client.TypeChatListMain {
					offsetOrder = int64(lastChat.Positions[i].Order)
				}
			}
			offsetChatId = lastChat.Id
		}
		chats, err := tdlibClient.GetChats(&client.GetChatsRequest{
			ChatList:     &client.ChatListMain{},
			Limit:        int32(limit - len(allChats)),
			OffsetOrder:  client.JsonInt64(offsetOrder),
			OffsetChatId: offsetChatId,
		})
		if err != nil {
			return nil, err
		}
		if len(chats.ChatIds) == 0 {
			return allChats, nil
		}
		for _, chatId := range chats.ChatIds {
			chat, err := tdlibClient.GetChat(&client.GetChatRequest{
				ChatId: chatId,
			})
			if err == nil {
				allChats = append(allChats, chat)
			} else {
				return nil, err
			}
		}
	}
	return allChats, nil
}

// OK: перенесено - service/media_album/service.go
type MediaAlbum struct {
	messages     []*client.Message
	lastReceived time.Time
}

var mediaAlbums = make(map[string]MediaAlbum)

// https://github.com/tdlib/td/issues/1482
// НЕТ: перенесено частично - service/media_album/service.go (AddMessage)
func addMessageToMediaAlbum(forwardKey string, message *client.Message) bool {
	key := fmt.Sprintf("%s:%d", forwardKey, message.MediaAlbumId)
	item, ok := mediaAlbums[key]
	if !ok {
		item = MediaAlbum{}
	}
	item.messages = append(item.messages, message)
	item.lastReceived = time.Now()
	mediaAlbums[key] = item
	return !ok
}

// НЕТ: перенесено частично - service/media_album/service.go (GetLastReceivedDiff)
func getMediaAlbumLastReceivedDiff(key string) time.Duration {
	mediaAlbumsMu.Lock()
	defer mediaAlbumsMu.Unlock()
	return time.Since(mediaAlbums[key].lastReceived)
}

// НЕТ: перенесено частично - service/media_album/service.go (GetMessages)
func getMediaAlbumMessages(key string) []*client.Message {
	mediaAlbumsMu.Lock()
	defer mediaAlbumsMu.Unlock()
	messages := mediaAlbums[key].messages
	delete(mediaAlbums, key)
	return messages
}

const waitForMediaAlbum = 3 * time.Second

// НЕТ: не перенесено, предлагаю - service/media_album/service.go (HandleMediaAlbum)
func handleMediaAlbum(forwardKey string, id client.JsonInt64, cb func(messages []*client.Message)) {
	key := fmt.Sprintf("%s:%d", forwardKey, id)
	diff := getMediaAlbumLastReceivedDiff(key)
	if diff < waitForMediaAlbum {
		time.Sleep(waitForMediaAlbum - diff)
		handleMediaAlbum(forwardKey, id, cb)
		return
	}
	messages := getMediaAlbumMessages(key)
	cb(messages)
}

// НЕТ: перенесено частично - service/engine/service.go (processMessage)
func doUpdateNewMessage(messages []*client.Message, forwardKey string, forward config.Forward, forwardedTo map[int64]bool, checkFns map[int64]func(), otherFns map[int64]func()) {
	src := messages[0]
	formattedText, contentMode := getFormattedText(src.Content)
	log.Printf("doUpdateNewMessage > do > ChatId: %d Id: %d hasText: %t MediaAlbumId: %d", src.ChatId, src.Id, formattedText.Text != "", src.MediaAlbumId)
	// for log
	var (
		isFilters = false
		isOther   = false
		result    []int64
	)
	defer func() {
		log.Printf("doUpdateNewMessage > ok > ChatId: %d Id: %d isFilters: %t isOther: %t result: %v", src.ChatId, src.Id, isFilters, isOther, result)
	}()
	if contentMode == "" {
		log.Print("contentMode == \"\"")
		return
	}
	switch checkFilters(formattedText, forward) {
	case FiltersOK:
		isFilters = true
		// checkFns[forward.Check] = nil // !! не надо сбрасывать - хочу проверить сообщение, даже если где-то прошли фильтры
		otherFns[forward.Other] = nil
		for _, dstChatId := range forward.To {
			if isNotForwardedTo(forwardedTo, dstChatId) {
				forwardNewMessages(tdlibClient, messages, src.ChatId, dstChatId, forward.SendCopy, forwardKey)
				result = append(result, dstChatId)
			}
		}
	case FiltersCheck:
		if forward.Check != 0 {
			_, ok := checkFns[forward.Check]
			if !ok {
				checkFns[forward.Check] = func() {
					const isSendCopy = false // обязательно надо форвардить, иначе невидно текущего сообщения
					forwardNewMessages(tdlibClient, messages, src.ChatId, forward.Check, isSendCopy, forwardKey)
				}
			}
		}
	case FiltersOther:
		if forward.Other != 0 {
			_, ok := otherFns[forward.Other]
			if !ok {
				otherFns[forward.Other] = func() {
					const isSendCopy = true // обязательно надо копировать, иначе невидно редактирование исходного сообщения
					forwardNewMessages(tdlibClient, messages, src.ChatId, forward.Other, isSendCopy, forwardKey)
				}
			}
		}
	}
}

// func getConfig() *config.Config {
// 	configMu.Lock()
// 	defer configMu.Unlock()
// 	result := configData // ???
// 	return result
// }

// NO: не перенесено, предлагаю - util/panic.go
func handlePanic() {
	if err := recover(); err != nil {
		log.Printf("Panic...\n%s\n\n%s", err, debug.Stack())
		os.Exit(1)
	}
}

const viewedMessagesPrefix = "viewedMsgs"

// НЕТ: перенесено частично - service/storage/service.go (IncrementViewedMessages)
func incrementViewedMessages(toChatId int64) {
	date := time.Now().UTC().Format("2006-01-02")
	key := []byte(fmt.Sprintf("%s:%d:%s", viewedMessagesPrefix, toChatId, date))
	val := incrementByDB(key)
	log.Printf("incrementViewedMessages > key: %s val: %d", key, int64(bytesToUint64(val)))
}

const forwardedMessagesPrefix = "forwardedMsgs"

// НЕТ: перенесено частично - service/storage/service.go (IncrementForwardedMessages)
func incrementForwardedMessages(toChatId int64) {
	date := time.Now().UTC().Format("2006-01-02")
	key := []byte(fmt.Sprintf("%s:%d:%s", forwardedMessagesPrefix, toChatId, date))
	val := incrementByDB(key)
	log.Printf("incrementForwardedMessages > key: %s val: %d", key, int64(bytesToUint64(val)))
}

var forwardedToMu sync.Mutex

// НЕТ: не перенесено, предлагаю - service/engine/service.go (IsNotForwardedTo)
func isNotForwardedTo(forwardedTo map[int64]bool, dstChatId int64) bool {
	forwardedToMu.Lock()
	defer forwardedToMu.Unlock()
	if !forwardedTo[dstChatId] {
		forwardedTo[dstChatId] = true
		return true
	}
	return false
}

// **** db routines

// OK: перенесено - repo/storage/repo.go (convertUint64ToBytes)
func uint64ToBytes(i uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], i)
	return buf[:]
}

// OK: перенесено - repo/storage/repo.go (ConvertBytesToUint64)
func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// OK: перенесено - repo/storage/repo.go (Increment)
func incrementByDB(key []byte) []byte {
	// Merge function to add two uint64 numbers
	add := func(existing, new []byte) []byte {
		return uint64ToBytes(bytesToUint64(existing) + bytesToUint64(new))
	}
	m := badgerDB.GetMergeOperator(key, add, 200*time.Millisecond)
	defer m.Stop()
	m.Add(uint64ToBytes(1))
	result, _ := m.Get()
	return result
}

// OK: перенесено - repo/storage/repo.go (Get)
func getForDB(key []byte) []byte {
	var (
		err error
		val []byte
	)
	err = badgerDB.View(func(txn *badger.Txn) error {
		var item *badger.Item
		item, err = txn.Get(key)
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Printf("getForDB > key: %s %s", key, err)
	} else {
		log.Printf("getForDB > key: %s, val: %s", key, string(val))
	}
	return val
}

// OK: перенесено - repo/storage/repo.go (Set)
func setForDB(key []byte, val []byte) {
	err := badgerDB.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, val)
		return err
	})
	if err != nil {
		log.Printf("setForDB > key: %s err: %s", string(key), err)
	} else {
		log.Printf("setForDB > key: %s val: %s", string(key), string(val))
	}
}

// OK: перенесено - repo/storage/repo.go (Delete)
func deleteForDB(key []byte) {
	err := badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
	if err != nil {
		log.Printf("deleteForDB > key: %s err: %s", string(key), err)
	}
}
