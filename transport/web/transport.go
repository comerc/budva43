package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/zelenin/go-tdlib/client"
	"go.uber.org/atomic"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/entity"
)

// messageController определяет интерфейс контроллера сообщений, необходимый для HTTP транспорта
type messageController interface {
	GetMessage(chatID, messageID int64) (*client.Message, error)
	SendMessage(chatID int64, text string) (*client.Message, error)
	DeleteMessage(chatID, messageID int64) error
	EditMessage(chatID, messageID int64, text string) (*client.Message, error)
	FormatMessage(text, fromFormat, toFormat string) (string, error)
	GetMessageText(message *client.Message) string
	GetContentType(message *client.Message) string
}

// forwardController определяет интерфейс контроллера пересылок, необходимый для HTTP транспорта
type forwardController interface {
	GetForwardRule(id string) (*entity.ForwardRule, error)
	SaveForwardRule(rule *entity.ForwardRule) error
}

// reportController определяет интерфейс контроллера отчетов, необходимый для HTTP транспорта
type reportController interface {
	GenerateActivityReport(startDate, endDate time.Time) (*entity.ActivityReport, error)
	GenerateForwardingReport(startDate, endDate time.Time) (*entity.ForwardingReport, error)
	GenerateErrorReport(startDate, endDate time.Time) (*entity.ErrorReport, error)
}

// Transport представляет HTTP маршрутизатор для API
type Transport struct {
	messageController messageController
	forwardController forwardController
	reportController  reportController
	server            *http.Server
	mux               *http.ServeMux
	isClosed          *atomic.Bool
}

// New создает новый экземпляр HTTP маршрутизатора
func New(
	messageController messageController,
	forwardController forwardController,
	reportController reportController,
) *Transport {
	return &Transport{
		messageController: messageController,
		forwardController: forwardController,
		reportController:  reportController,
		isClosed:          atomic.NewBool(false),
	}
}

// SetupRoutes настраивает HTTP маршруты
func (r *Transport) SetupRoutes(mux *http.ServeMux) {
	// Маршруты для сообщений
	mux.HandleFunc("/api/messages", r.handleMessages)
	mux.HandleFunc("/api/messages/", r.handleMessageByID)

	// Маршруты для правил пересылки
	mux.HandleFunc("/api/forward-rules", r.handleForwardRules)
	mux.HandleFunc("/api/forward-rules/", r.handleForwardRuleByID)

	// Маршруты для отчетов
	mux.HandleFunc("/api/reports", r.handleReports)

	// Маршрут для основной страницы
	mux.HandleFunc("/", r.handleRoot)
}

// handleRoot обрабатывает запросы к корневому маршруту
func (r *Transport) handleRoot(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Budva43 API Server"))
}

// handleMessages обрабатывает запросы для работы с сообщениями
func (r *Transport) handleMessages(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		// Получение списка сообщений - не реализовано
		http.Error(w, "Not implemented", http.StatusNotImplemented)

	case http.MethodPost:
		// Отправка нового сообщения
		var messageRequest struct {
			ChatID int64  `json:"chat_id"`
			Text   string `json:"text"`
		}
		if err := json.NewDecoder(req.Body).Decode(&messageRequest); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		message, err := r.messageController.SendMessage(messageRequest.ChatID, messageRequest.Text)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error sending message: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(message)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleMessageByID обрабатывает запросы для работы с конкретным сообщением
func (r *Transport) handleMessageByID(w http.ResponseWriter, req *http.Request) {
	// Извлекаем ID сообщения из URL
	path := req.URL.Path
	if len(path) <= len("/api/messages/") {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// Получаем параметры
	messageIDStr := path[len("/api/messages/"):]
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	chatIDStr := req.URL.Query().Get("chat_id")
	if chatIDStr == "" {
		http.Error(w, "Missing chat_id parameter", http.StatusBadRequest)
		return
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodGet:
		// Получение сообщения
		message, err := r.messageController.GetMessage(chatID, messageID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting message: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(message)

	case http.MethodPut:
		// Редактирование сообщения
		var requestBody struct {
			Text string `json:"text"`
		}

		if err := json.NewDecoder(req.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		result, err := r.messageController.EditMessage(chatID, messageID, requestBody.Text)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error editing message: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)

	case http.MethodDelete:
		// Удаление сообщения
		err := r.messageController.DeleteMessage(chatID, messageID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error deleting message: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleForwardRules обрабатывает запросы для работы с правилами пересылки
func (r *Transport) handleForwardRules(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		// Создание нового правила пересылки
		var rule entity.ForwardRule
		if err := json.NewDecoder(req.Body).Decode(&rule); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		err := r.forwardController.SaveForwardRule(&rule)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error saving forward rule: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleForwardRuleByID обрабатывает запросы для работы с конкретным правилом пересылки
func (r *Transport) handleForwardRuleByID(w http.ResponseWriter, req *http.Request) {
	// Извлекаем ID правила из URL
	path := req.URL.Path
	if len(path) <= len("/api/forward-rules/") {
		http.Error(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}

	ruleID := path[len("/api/forward-rules/"):]

	switch req.Method {
	case http.MethodGet:
		// Получение правила пересылки
		rule, err := r.forwardController.GetForwardRule(ruleID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting forward rule: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rule)

	case http.MethodPut:
		// Обновление правила пересылки
		var rule entity.ForwardRule
		if err := json.NewDecoder(req.Body).Decode(&rule); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Устанавливаем ID из URL
		rule.ID = ruleID

		err := r.forwardController.SaveForwardRule(&rule)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating forward rule: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleReports обрабатывает запросы для работы с отчетами
func (r *Transport) handleReports(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметры запроса
	reportType := req.URL.Query().Get("type")
	if reportType == "" {
		http.Error(w, "Missing type parameter", http.StatusBadRequest)
		return
	}

	startDateStr := req.URL.Query().Get("start_date")
	endDateStr := req.URL.Query().Get("end_date")

	// Парсим даты
	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -7) // По умолчанию - неделя назад
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	} else {
		endDate = time.Now() // По умолчанию - текущая дата
	}

	var report interface{}

	// Генерируем отчет в зависимости от типа
	switch reportType {
	case "activity":
		report, err = r.reportController.GenerateActivityReport(startDate, endDate)
	case "forwarding":
		report, err = r.reportController.GenerateForwardingReport(startDate, endDate)
	case "error":
		report, err = r.reportController.GenerateErrorReport(startDate, endDate)
	default:
		http.Error(w, "Invalid report type. Use 'activity', 'forwarding', or 'error'", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating report: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// Start запускает HTTP-сервер
func (t *Transport) Start(ctx context.Context) error {
	// Создаем новый мультиплексор
	t.mux = http.NewServeMux()

	// Настраиваем маршруты
	t.SetupRoutes(t.mux)

	// Настраиваем HTTP-сервер
	t.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Web.Host, config.Web.Port),
		Handler:      t.mux,
		ReadTimeout:  config.Web.ReadTimeout,
		WriteTimeout: config.Web.WriteTimeout,
	}

	// Запускаем HTTP-сервер в отдельной горутине
	go func() {
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server terminated with error", "err", err)
		}
	}()

	slog.Info("HTTP server started", "addr", t.server.Addr)

	return nil
}

// Stop останавливает HTTP-сервер
func (t *Transport) Stop() error {
	if t.isClosed.Swap(true) || t.server == nil {
		return nil
	}

	slog.Info("Stopping HTTP server")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), config.Web.ShutdownTimeout)
	defer cancel()

	if err := t.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down HTTP server: %w", err)
	}

	slog.Info("HTTP server stopped")
	return nil
}
