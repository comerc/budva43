package http

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Config представляет конфигурацию HTTP-сервера
type Config struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// Server представляет HTTP-сервер для API
type Server struct {
	server   *http.Server
	router   *Router
	config   Config
	mux      *http.ServeMux
	isClosed bool
}

// NewServer создает новый экземпляр HTTP-сервера
func NewServer(
	router *Router,
	config Config,
) *Server {
	mux := http.NewServeMux()

	// Настраиваем HTTP-сервер
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      mux,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return &Server{
		server:   server,
		router:   router,
		config:   config,
		mux:      mux,
		isClosed: false,
	}
}

// Start запускает HTTP-сервер
func (s *Server) Start(ctx context.Context) error {
	// Настраиваем маршруты
	s.router.SetupRoutes(s.mux)

	// Добавляем обработчик для основной страницы
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Budva43 API Server"))
	})

	// Запускаем HTTP-сервер в отдельной горутине
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting HTTP server: %v\n", err)
		}
	}()

	fmt.Printf("HTTP server started on %s\n", s.server.Addr)

	// Ожидаем сигнал остановки через контекст
	<-ctx.Done()

	return s.Stop()
}

// Stop останавливает HTTP-сервер
func (s *Server) Stop() error {
	if s.isClosed {
		return nil
	}

	s.isClosed = true
	fmt.Println("Shutting down HTTP server...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down HTTP server: %w", err)
	}

	fmt.Println("HTTP server stopped")
	return nil
}
