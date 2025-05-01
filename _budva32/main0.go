package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/comerc/budva32/config"
	"github.com/dgraph-io/badger"
	"github.com/joho/godotenv"
	"github.com/zelenin/go-tdlib/client"
)

const (
	projectName = "budva32"
)

var (
	inputCh  = make(chan string, 1)
	outputCh = make(chan string, 1)
	//
	uniqueFrom map[int64]struct{}
	//
	configData    *config.Config
	tdlibClient   *client.Client
	mediaAlbumsMu sync.Mutex
	// configMu      sync.Mutex
	badgerDB *badger.DB
)

func main() {
	// OK: перенесено - main.go (setupLogger)
	log.SetFlags(log.LUTC | log.Ldate | log.Ltime | log.Lshortfile)

	var err error

	// OK: перенесено - config/loader.go (load)
	if err = godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	// OK: перенесено - config/dir.go (MakeDirs)
	path := filepath.Join(".", ".tdata")
	if _, err = os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	{
		// OK: перенесено - config/dir.go (MakeDirs)
		path := filepath.Join(path, "badger")
		if _, err = os.Stat(path); os.IsNotExist(err) {
			os.Mkdir(path, os.ModePerm)
		}
		// OK: перенесено - repo/badger/repo.go (Start)
		badgerDB, err = badger.Open(badger.DefaultOptions(path))
		if err != nil {
			log.Fatal(err)
		}
	}
	// OK: перенесено - repo/badger/repo.go (Close)
	defer badgerDB.Close()

	// OK: перенесено - repo/badger/repo.go (runGarbageCollection)
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
		again:
			err := badgerDB.RunValueLogGC(0.7)
			if err == nil {
				goto again
			}
		}
	}()

	// OK: перенесено - config/loader.go (load)
	var (
		apiId       = os.Getenv("BUDVA32_API_ID")
		apiHash     = os.Getenv("BUDVA32_API_HASH")
		phonenumber = os.Getenv("BUDVA32_PHONENUMBER")
		port        = os.Getenv("BUDVA32_PORT")
	)

	reload := func() {
		// OK: перенесено - config/loader.go (load)
		tmpConfigData, err := config.Load()
		if err != nil {
			log.Fatalf("Can't initialise config: %s", err)
			return
		}
		// OK: перенесено - service/engine/service.go (validateConfig)
		for _, v := range tmpConfigData.ReplaceFragments {
			for from, to := range v {
				if strLenForUTF16(from) != strLenForUTF16(to) {
					err := fmt.Errorf(`strLen("%s") != strLen("%s")`, from, to)
					log.Print(err)
					return
				}
			}
		}
		// OK: перенесено - service/engine/service.go (enrichConfig)
		tmpUniqueFrom := make(map[int64]struct{})
		re := regexp.MustCompile("[:,]")
		for forwardKey, forward := range tmpConfigData.Forwards {
			if re.FindString(forwardKey) != "" {
				err := fmt.Errorf("cannot use [:,] as Config.Forwards key in %s", forwardKey)
				log.Print(err)
				return
			}
			// TODO: "destination Id cannot be equal to source Id" - для всех From-To,
			// а не только для одного Forward; для будущей обработки To в UpdateDeleteMessages
			for _, dscChatId := range forward.To {
				if forward.From == dscChatId {
					err := fmt.Errorf("destination Id cannot be equal to source Id %d", dscChatId)
					log.Print(err)
					return
				}
			}
			tmpUniqueFrom[forward.From] = struct{}{}
		}
		// configMu.Lock()
		// defer configMu.Unlock()
		uniqueFrom = tmpUniqueFrom
		configData = tmpConfigData
	}

	// НЕТ: перенесено частично - config/config.go (Watch)
	go config.Watch(reload)

	// НЕТ: не перенесено, предлагаю - transport/web/transport.go (StartHTTPServer)
	go func() {
		http.HandleFunc("/favicon.ico", getFaviconHandler)
		http.HandleFunc("/", withBasicAuth(withAuthentiation(getChatsHandler)))
		http.HandleFunc("/ping", getPingHandler)
		http.HandleFunc("/answer", getAnswerHandler)
		host := getIP()
		port := ":" + port
		fmt.Println("Web-server is running: http://" + host + port)
		if err := http.ListenAndServe(port, http.DefaultServeMux); err != nil {
			log.Fatal("Error starting http server: ", err)
			return
		}
	}()

	// НЕТ: не перенесено, предлагаю - service/auth/service.go (RequestAuthData)
	// client authorizer
	authorizer := client.ClientAuthorizer()
	go func() {
		for {
			state, ok := <-authorizer.State
			if !ok {
				return
			}
			switch state.AuthorizationStateType() {
			case client.TypeAuthorizationStateWaitPhoneNumber:
				authorizer.PhoneNumber <- phonenumber
			case client.TypeAuthorizationStateWaitCode:
				outputCh <- fmt.Sprintf("Enter code for %s: ", phonenumber)
				code := <-inputCh
				authorizer.Code <- code
			case client.TypeAuthorizationStateWaitPassword:
				outputCh <- fmt.Sprintf("Enter password for %s: ", phonenumber)
				password := <-inputCh
				authorizer.Password <- password
			case client.TypeAuthorizationStateReady:
				return
			}
		}
	}()

	// or bot authorizer
	// botToken := "000000000:gsVCGG5YbikxYHC7bP5vRvmBqJ7Xz6vG6td"
	// authorizer := client.BotAuthorizer(botToken)

	// НЕТ: не перенесено, предлагаю - service/telegram/client.go (InitClient)
	authorizer.TdlibParameters <- &client.TdlibParameters{
		UseTestDc:              false,
		DatabaseDirectory:      filepath.Join(path, "db"),
		FilesDirectory:         filepath.Join(path, "files"),
		UseFileDatabase:        false,
		UseChatInfoDatabase:    false,
		UseMessageDatabase:     true,
		UseSecretChats:         false,
		ApiId:                  int32(convertToInt(apiId)),
		ApiHash:                apiHash,
		SystemLanguageCode:     "en",
		DeviceModel:            "Server",
		SystemVersion:          "1.0.0",
		ApplicationVersion:     "1.0.0",
		EnableStorageOptimizer: true,
		IgnoreFileNames:        false,
	}

	// НЕТ: не перенесено, предлагаю - service/telegram/client.go (SetupLogs)
	logStream := func(tdlibClient *client.Client) {
		tdlibClient.SetLogStream(&client.SetLogStreamRequest{
			LogStream: &client.LogStreamFile{
				Path:           filepath.Join(path, ".log"),
				MaxFileSize:    10485760,
				RedirectStderr: true,
			},
		})
	}

	logVerbosity := func(tdlibClient *client.Client) {
		tdlibClient.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
			NewVerbosityLevel: 1,
		})
	}

	// НЕТ: не перенесено, предлагаю - service/telegram/client.go (NewClient)
	tdlibClient, err = client.NewClient(authorizer, logStream, logVerbosity)
	if err != nil {
		log.Fatalf("NewClient error: %s", err)
	}
	defer tdlibClient.Stop()

	outputCh <- "Ready!"

	log.Print("Start...")

	// НЕТ: не перенесено, предлагаю - service/telegram/client.go (GetVersion)
	if optionValue, err := tdlibClient.GetOption(&client.GetOptionRequest{
		Name: "version",
	}); err != nil {
		log.Fatalf("GetOption error: %s", err)
	} else {
		log.Printf("TDLib version: %s", optionValue.(*client.OptionValueString).Value)
	}

	// НЕТ: не перенесено, предлагаю - service/telegram/user.go (GetMe)
	if me, err := tdlibClient.GetMe(); err != nil {
		log.Fatalf("GetMe error: %s", err)
	} else {
		log.Printf("Me: %s %s [@%s]", me.FirstName, me.LastName, me.Username)
	}

	// НЕТ: не перенесено, предлагаю - service/telegram/client.go (GetListener)
	listener := tdlibClient.GetListener()
	defer listener.Close()

	// НЕТ: не перенесено, предлагаю - internal/app/signals.go (HandleSignals)
	// Handle Ctrl+C
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		log.Print("Stop...")
		os.Exit(1)
	}()

	// NO: не перенесено, предлагаю - util/panic.go (HandlePanic)
	defer handlePanic()

	// НЕТ: не перенесено, предлагаю - service/report/service.go (StartReportService)
	go runReports()

	// OK: перенесено - service/queue/service.go (Start)
	go runQueue()

	// НЕТ: перенесено частично - service/engine/service.go (handleUpdates)
	for update := range listener.Updates {
		handleUpdate(update)
	}
}
