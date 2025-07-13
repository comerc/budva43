package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/transport/web/mocks"
)

func Test(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		authState            client.AuthorizationState
		expectedStateType    string
		expectedPasswordHint string
	}{
		{
			name: "with password hint",
			authState: &client.AuthorizationStateWaitPassword{
				PasswordHint: "test hint",
			},
			expectedStateType:    "authorizationStateWaitPassword",
			expectedPasswordHint: "test hint",
		},
		{
			name: "without password hint",
			authState: &client.AuthorizationStateWaitPassword{
				PasswordHint: "",
			},
			expectedStateType:    "authorizationStateWaitPassword",
			expectedPasswordHint: "",
		},
		{
			name:                 "not password state - wait code",
			authState:            &client.AuthorizationStateWaitCode{},
			expectedStateType:    "authorizationStateWaitCode",
			expectedPasswordHint: "",
		},
		{
			name:                 "not password state - wait phone",
			authState:            &client.AuthorizationStateWaitPhoneNumber{},
			expectedStateType:    "authorizationStateWaitPhoneNumber",
			expectedPasswordHint: "",
		},
		{
			name:                 "nil state",
			authState:            nil,
			expectedStateType:    "",
			expectedPasswordHint: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			authService := mocks.NewAuthService(t)
			facadeGQL := mocks.NewFacadeGQL(t)

			transport := New(authService, facadeGQL)
			transport.authState = test.authState

			req := httptest.NewRequest(http.MethodGet, "/api/auth/telegram/state", nil)
			w := httptest.NewRecorder()

			transport.handleAuthState(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]any
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStateType, response["state_type"])

			passwordHint, hasPasswordHint := response["password_hint"]
			if test.expectedPasswordHint == "" {
				assert.False(t, hasPasswordHint, "password_hint should not be present")
			} else {
				assert.True(t, hasPasswordHint, "password_hint should be present")
				assert.Equal(t, test.expectedPasswordHint, passwordHint)
			}
		})
	}
}
