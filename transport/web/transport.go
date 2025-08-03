package web

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/dto/gql/dto"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
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

//go:generate mockery --name=facadeGQL --exported
type facadeGQL interface {
	GetStatus() (*dto.Status, error)
}

// Transport представляет HTTP маршрутизатор для API
type Transport struct {
	log *log.Logger
	//
	authService authService
	facadeGQL   facadeGQL
	authState   client.AuthorizationState
	server      *http.Server
}

// New создает новый экземпляр HTTP маршрутизатора
func New(
	authService authService,
	facadeGQL facadeGQL,
) *Transport {
	return &Transport{
		log: log.NewLogger(),
		//
		authService: authService,
		facadeGQL:   facadeGQL,
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

// StartContext запускает HTTP-сервер
func (t *Transport) StartContext(ctx context.Context, shutdown func()) error {
	_ = shutdown // пока не используется

	addr := net.JoinHostPort(config.Web.Host, config.Web.Port)
	if !util.IsPortFree(addr) {
		return log.NewError(
			fmt.Sprintf("port is busy -> task kill-port -- %s", config.Grpc.Port),
			"addr", addr,
		)
	}

	t.authService.Subscribe(t.newFuncNotify())

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
func (t *Transport) newFuncNotify() notify {
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

	mux.HandleFunc("/graphql", newFuncHandleGraph(t.facadeGQL))
	mux.HandleFunc("/playground", playground.Handler("GraphQL playground", "/graphql"))

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
