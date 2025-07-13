package web

import (
	"encoding/json"
	"net/http"

	"github.com/zelenin/go-tdlib/client"
)

// handleAuthState обработчик для получения текущего состояния авторизации
func (t *Transport) handleAuthState(w http.ResponseWriter, r *http.Request) {
	var err error
	defer t.log.ErrorOrDebug(&err, "")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var stateType string
	var passwordHint string
	state := t.authState

	if state != nil {
		stateType = state.AuthorizationStateType()

		// Если состояние - ожидание пароля, извлекаем подсказку
		if stateType == client.TypeAuthorizationStateWaitPassword {
			passwordState := state.(*client.AuthorizationStateWaitPassword)
			passwordHint = passwordState.PasswordHint
		}
	}

	response := map[string]any{
		"state_type": stateType,
	}

	// Добавляем подсказку пароля только если она есть
	if passwordHint != "" {
		response["password_hint"] = passwordHint
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
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
