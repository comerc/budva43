package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/entity"
)

// TODO: реализовать авторизацию telegram-клиента для web-транспорта

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

// authTelegramController определяет интерфейс контроллера авторизации Telegram
type authTelegramController interface {
	SubmitPhoneNumber(phone string)
	SubmitCode(code string)
	SubmitPassword(password string)
	GetAuthorizationState() (client.AuthorizationState, error)
}

// Transport представляет HTTP маршрутизатор для API
type Transport struct {
	log               *slog.Logger
	messageController messageController
	forwardController forwardController
	reportController  reportController
	authController    authTelegramController
	authClients       map[string]chan client.AuthorizationState
	server            *http.Server
}

// New создает новый экземпляр HTTP маршрутизатора
func New(
	messageController messageController,
	forwardController forwardController,
	reportController reportController,
	authController authTelegramController,
) *Transport {
	return &Transport{
		log: slog.With("module", "transport.web"),

		messageController: messageController,
		forwardController: forwardController,
		reportController:  reportController,
		authController:    authController,
		authClients:       make(map[string]chan client.AuthorizationState),
	}
}

// setupRoutes настраивает HTTP маршруты
func (t *Transport) setupRoutes(mux *http.ServeMux) {
	// Маршруты для сообщений
	mux.HandleFunc("/api/messages", t.handleMessages)
	mux.HandleFunc("/api/messages/", t.handleMessageByID)

	// Маршруты для правил пересылки
	mux.HandleFunc("/api/forward-rules", t.handleForwardRules)
	mux.HandleFunc("/api/forward-rules/", t.handleForwardRuleByID)

	// Маршруты для отчетов
	mux.HandleFunc("/api/reports", t.handleReports)

	// Маршруты для авторизации Telegram
	mux.HandleFunc("/api/auth/telegram/state", t.handleAuthState)
	mux.HandleFunc("/api/auth/telegram/phone", t.handleSubmitPhone)
	mux.HandleFunc("/api/auth/telegram/code", t.handleSubmitCode)
	mux.HandleFunc("/api/auth/telegram/password", t.handleSubmitPassword)
	mux.HandleFunc("/api/auth/telegram/events", t.handleAuthEvents)

	// Маршрут для основной страницы
	mux.HandleFunc("/", t.handleRoot)
}

// handleRoot обрабатывает запросы к корневому маршруту
func (t *Transport) handleRoot(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Budva43 API Server"))
}

func (t *Transport) logHandler(errPointer *error, now time.Time, name string) {
	err := *errPointer
	if err == nil {
		t.log.Info(name,
			"took", time.Since(now),
		)
	} else {
		t.log.Error(name,
			"took", time.Since(now),
			"err", err,
		)
	}
}

// handleMessages обрабатывает запросы для работы с сообщениями
func (t *Transport) handleMessages(w http.ResponseWriter, req *http.Request) {
	var err error
	defer t.logHandler(&err, time.Now(), "handleMessages")

	switch req.Method {
	case http.MethodGet:
		// Получение списка сообщений - не реализовано
		http.Error(w, "Not implemented", http.StatusNotImplemented)
		err = fmt.Errorf("not implemented")

	case http.MethodPost:
		// Отправка нового сообщения
		var messageRequest struct {
			ChatID int64  `json:"chat_id"`
			Text   string `json:"text"`
		}
		err = json.NewDecoder(req.Body).Decode(&messageRequest)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var message *client.Message
		message, err = t.messageController.SendMessage(messageRequest.ChatID, messageRequest.Text)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error sending message: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(message)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		err = fmt.Errorf("method not allowed")
	}
}

// handleMessageByID обрабатывает запросы для работы с конкретным сообщением
func (t *Transport) handleMessageByID(w http.ResponseWriter, req *http.Request) {
	var err error
	defer t.logHandler(&err, time.Now(), "handleMessages")

	// Получаем параметры
	query := req.URL.Query()

	messageIDStr := query.Get("message_id")
	var messageID int64
	messageID, err = strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid message_id", http.StatusBadRequest)
		return
	}

	chatIDStr := query.Get("chat_id")
	var chatID int64
	chatID, err = strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodGet:
		// Получение сообщения
		var message *client.Message
		message, err = t.messageController.GetMessage(chatID, messageID)
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

		err = json.NewDecoder(req.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var result *client.Message
		result, err = t.messageController.EditMessage(chatID, messageID, requestBody.Text)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error editing message: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)

	case http.MethodDelete:
		// Удаление сообщения
		err = t.messageController.DeleteMessage(chatID, messageID)
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
func (t *Transport) handleForwardRules(w http.ResponseWriter, req *http.Request) {
	var err error
	defer t.logHandler(&err, time.Now(), "handleMessages")

	switch req.Method {
	case http.MethodPost:
		// Создание нового правила пересылки
		var rule entity.ForwardRule
		err = json.NewDecoder(req.Body).Decode(&rule)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		err = t.forwardController.SaveForwardRule(&rule)
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
func (t *Transport) handleForwardRuleByID(w http.ResponseWriter, req *http.Request) {
	var err error
	defer t.logHandler(&err, time.Now(), "handleMessages")

	// Получаем параметры
	query := req.URL.Query()

	ruleID := query.Get("rule_id")
	if ruleID == "" {
		http.Error(w, "Missing rule_id parameter", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodGet:
		// Получение правила пересылки
		var rule *entity.ForwardRule
		rule, err = t.forwardController.GetForwardRule(ruleID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting forward rule: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rule)

	case http.MethodPut:
		// Обновление правила пересылки
		var rule entity.ForwardRule
		err = json.NewDecoder(req.Body).Decode(&rule)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Устанавливаем ID из URL
		rule.ID = ruleID

		err = t.forwardController.SaveForwardRule(&rule)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating forward rule: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK) // TODO: и так устанавливается по умолчанию?

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleReports обрабатывает запросы для работы с отчетами
func (t *Transport) handleReports(w http.ResponseWriter, req *http.Request) {
	var err error
	defer t.logHandler(&err, time.Now(), "handleMessages")

	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметры запроса
	query := req.URL.Query()

	reportType := query.Get("type")
	if reportType == "" {
		http.Error(w, "Missing type parameter", http.StatusBadRequest)
		return
	}

	startDateStr := query.Get("start_date")
	endDateStr := query.Get("end_date")

	// Парсим даты
	var startDate, endDate time.Time

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

	var report any

	// Генерируем отчет в зависимости от типа
	switch reportType {
	case "activity":
		report, err = t.reportController.GenerateActivityReport(startDate, endDate)
	case "forwarding":
		report, err = t.reportController.GenerateForwardingReport(startDate, endDate)
	case "error":
		report, err = t.reportController.GenerateErrorReport(startDate, endDate)
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

// handleAuthState обработчик для получения текущего состояния авторизации
func (t *Transport) handleAuthState(w http.ResponseWriter, r *http.Request) {
	var err error
	defer t.logHandler(&err, time.Now(), "handleMessages")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var state client.AuthorizationState
	state, err = t.authController.GetAuthorizationState()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting authorization state: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"state_type": state.AuthorizationStateType(),
	})
}

// handleSubmitPhone обработчик для отправки номера телефона
func (t *Transport) handleSubmitPhone(w http.ResponseWriter, r *http.Request) {
	var err error
	defer t.logHandler(&err, time.Now(), "handleMessages")

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

	t.authController.SubmitPhoneNumber(data.Phone)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]any{
		"status": "accepted",
	})
}

// handleSubmitCode обработчик для отправки кода подтверждения
func (t *Transport) handleSubmitCode(w http.ResponseWriter, r *http.Request) {
	var err error
	defer t.logHandler(&err, time.Now(), "handleMessages")

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

	t.authController.SubmitCode(data.Code)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]any{
		"status": "accepted",
	})
}

// handleSubmitPassword обработчик для отправки пароля
func (t *Transport) handleSubmitPassword(w http.ResponseWriter, r *http.Request) {
	var err error
	defer t.logHandler(&err, time.Now(), "handleMessages")

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

	t.authController.SubmitPassword(data.Password)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]any{
		"status": "accepted",
	})
}

// TODO: under construction
// handleAuthEvents устанавливает SSE соединение для получения обновлений состояния авторизации
func (t *Transport) handleAuthEvents(w http.ResponseWriter, r *http.Request) {
	var err error
	defer t.logHandler(&err, time.Now(), "handleMessages")

	// Настройка SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Создаем канал для событий авторизации
	clientID := generateClientID()
	events := make(chan client.AuthorizationState, 10)

	// Регистрируем клиента
	t.authClients[clientID] = events

	// Отправляем текущее состояние сразу при подключении
	var state client.AuthorizationState
	state, err = t.authController.GetAuthorizationState()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting authorization state: %v", err), http.StatusInternalServerError)
		return
	}
	if state != nil {
		events <- state
	}

	// Отправляем события клиенту
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Закрываем канал и удаляем клиента при завершении запроса
	defer func() {
		delete(t.authClients, clientID)
		close(events)
	}()

	// Обрабатываем закрытие соединения
	disconnect := r.Context().Done()

	for {
		select {
		case <-disconnect:
			// Клиент отключился
			return
		case state, ok := <-events:
			if !ok {
				// Канал закрыт
				return
			}
			// Отправляем событие клиенту
			fmt.Fprintf(w, "data: {\"state_type\": \"%s\"}\n\n", state.AuthorizationStateType())
			flusher.Flush()
		}
	}
}

// generateClientID генерирует уникальный идентификатор клиента
func generateClientID() string {
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}

// Start запускает HTTP-сервер
func (t *Transport) Start(ctx context.Context, shutdown func()) error {
	// Создаем новый мультиплексор
	mux := http.NewServeMux()

	// Настраиваем маршруты
	t.setupRoutes(mux)

	// Настраиваем HTTP-сервер
	t.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Web.Host, config.Web.Port),
		Handler:      mux,
		ReadTimeout:  config.Web.ReadTimeout,
		WriteTimeout: config.Web.WriteTimeout,
	}

	// Запускаем HTTP-сервер в отдельной горутине
	go func() {
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.log.Error("HTTP server terminated with error", "err", err)
		}
	}()

	t.log.Info("HTTP server started", "addr", t.server.Addr)

	return nil
}

// Stop останавливает HTTP-сервер
func (t *Transport) Stop() error {
	t.log.Info("Stopping HTTP server")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), config.Web.ShutdownTimeout)
	defer cancel()

	if err := t.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down HTTP server: %w", err)
	}

	t.log.Info("HTTP server stopped")
	return nil
}

// TODO: under construction
// OnAuthStateChanged обработчик изменения состояния авторизации
func (t *Transport) OnAuthStateChanged(state client.AuthorizationState) {
	t.log.Debug("Web транспорт получил обновление состояния авторизации",
		"state", state.AuthorizationStateType())

	// Отправляем обновление всем подключенным клиентам
	for clientID, clientChan := range t.authClients {
		select {
		case clientChan <- state:
			t.log.Debug("Отправлено обновление состояния клиенту", "clientID", clientID)
		default:
			t.log.Debug("Канал клиента заполнен, пропускаем обновление", "clientID", clientID)
		}
	}
}
