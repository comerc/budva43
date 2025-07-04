package web

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/log"
)

// TODO: передавать состояние авторизации в режиме SSE
// TODO: передавать статус успешной авторизации

type notify = func(state client.AuthorizationState)

//go:generate mockery --name=authService --exported
type authService interface {
	Subscribe(notify)
	GetInputChan() chan<- string
	// GetClientDone() <-chan any
	// GetStatus() string
}

// Transport представляет HTTP маршрутизатор для API
type Transport struct {
	log *log.Logger
	//
	authService authService
	authState   client.AuthorizationState
	server      *http.Server
}

// New создает новый экземпляр HTTP маршрутизатора
func New(
	authService authService,
) *Transport {
	return &Transport{
		log: log.NewLogger(),
		//
		authService: authService,
	}
}

// logMiddleware добавляет логирование времени выполнения запросов
func (t *Transport) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		defer func() {
			t.log.ErrorOrDebug(nil, "",
				"method", r.Method,
				"path", r.URL.Path,
				"took", time.Since(now),
			)
		}()

		next.ServeHTTP(w, r)
	})
}

// Start запускает HTTP-сервер
func (t *Transport) Start(ctx context.Context, shutdown func()) error {
	_ = shutdown // не используется

	addr := net.JoinHostPort(config.Web.Host, config.Web.Port)
	if !isPortFree(addr) {
		return log.NewError("port is busy -> make kill-port", "addr", addr)
	}

	t.authService.Subscribe(newFuncNotify(t))

	t.createServer()

	go t.runServer()

	return nil
}

// Close останавливает HTTP-сервер
func (t *Transport) Close() error {
	var err error

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), config.Web.ShutdownTimeout)
	defer cancel()

	err = t.server.Shutdown(ctx)
	if err != nil {
		return log.WrapError(err) // внешняя ошибка
	}

	return nil
}

// newFuncNotify создает функцию для отправки состояния авторизации
func newFuncNotify(t *Transport) notify {
	return func(state client.AuthorizationState) {
		t.authState = state
	}
}

func (t *Transport) createServer() {
	// Создаем новый мультиплексор
	mux := http.NewServeMux()

	// Настраиваем маршруты
	t.setupRoutes(mux)

	// Оборачиваем весь мультиплексор в middleware для логирования
	handler := t.logMiddleware(mux)

	// Настраиваем HTTP-сервер
	t.server = &http.Server{
		Addr:         net.JoinHostPort(config.Web.Host, config.Web.Port),
		Handler:      handler,
		ReadTimeout:  config.Web.ReadTimeout,
		WriteTimeout: config.Web.WriteTimeout,
	}
}

func (t *Transport) runServer() {
	var err error
	defer t.log.ErrorOrDebug(&err, "", "addr", t.server.Addr)

	err = t.server.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
}

// setupRoutes настраивает HTTP маршруты
func (t *Transport) setupRoutes(mux *http.ServeMux) {
	// TODO: перенести в middleware?
	mux.HandleFunc("/api/auth/telegram/state", t.handleAuthState)
	mux.HandleFunc("/api/auth/telegram/phone", t.handleSubmitPhone)
	mux.HandleFunc("/api/auth/telegram/code", t.handleSubmitCode)
	mux.HandleFunc("/api/auth/telegram/password", t.handleSubmitPassword)

	mux.HandleFunc("/favicon.ico", t.handleFavicon)
	mux.HandleFunc("/", t.handleRoot)
}

// handleFavicon обрабатывает запросы к favicon
func (t *Transport) handleFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "app/static/favicon.ico")
}

// handleRoot обрабатывает запросы к корневому маршруту
func (t *Transport) handleRoot(w http.ResponseWriter, _ *http.Request) {
	var err error
	defer t.log.ErrorOrDebug(&err, "")

	w.Header().Set("Content-Type", "text/plain")
	_, err = w.Write([]byte("Budva43 API Server"))
}

// handleAuthState обработчик для получения текущего состояния авторизации
func (t *Transport) handleAuthState(w http.ResponseWriter, r *http.Request) {
	var err error
	defer t.log.ErrorOrDebug(&err, "")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var stateType string
	state := t.authState

	if state != nil {
		stateType = state.AuthorizationStateType()
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]any{
		"state_type": stateType,
	})
}

// handleSubmitPhone обработчик для отправки номера телефона
func (t *Transport) handleSubmitPhone(w http.ResponseWriter, r *http.Request) {
	var err error
	defer t.log.ErrorOrDebug(&err, "")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Phone string `json:"phone"`
	}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	t.authService.GetInputChan() <- data.Phone

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(map[string]any{
		"status": "accepted",
	})
}

// handleSubmitCode обработчик для отправки кода подтверждения
func (t *Transport) handleSubmitCode(w http.ResponseWriter, r *http.Request) {
	var err error
	defer t.log.ErrorOrDebug(&err, "")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Code string `json:"code"`
	}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	t.authService.GetInputChan() <- data.Code

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(map[string]any{
		"status": "accepted",
	})
}

// handleSubmitPassword обработчик для отправки пароля
func (t *Transport) handleSubmitPassword(w http.ResponseWriter, r *http.Request) {
	var err error
	defer t.log.ErrorOrDebug(&err, "")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Password string `json:"password"`
	}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	t.authService.GetInputChan() <- data.Password

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(map[string]any{
		"status": "accepted",
	})
}

// isPortFree проверяет, свободен ли порт
func isPortFree(addr string) bool {
	// Пытаемся подключиться к порту как клиент
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		// Если не удалось подключиться, порт свободен
		return true
	}
	// Если подключились, значит порт занят
	conn.Close()
	return false
}
