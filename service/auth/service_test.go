package auth

import (
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/service/auth/mocks"
)

func TestAuthService_GetStatus(t *testing.T) {
	t.Parallel()

	telegramRepo := mocks.NewTelegramRepo(t)
	telegramRepo.EXPECT().GetOption(&client.GetOptionRequest{Name: "version"}).
		Return(&client.OptionValueString{Value: "1.2.3"}, nil)
	telegramRepo.EXPECT().GetMe().
		Return(&client.User{Id: 123}, nil)
	service := New(telegramRepo)

	status := service.GetStatus()
	assert.Equal(t, "TDLib version: 1.2.3 userId: 123", status)
}

func TestAuthService_Subscribe(t *testing.T) {
	t.Parallel()

	telegramRepo := mocks.NewTelegramRepo(t)
	service := New(telegramRepo)

	// Проверяем начальное состояние
	require.Equal(t, 0, len(service.subscribers))

	// Создаем мок функции notify с помощью замыкания
	var called1, called2 bool
	mockNotify1 := func(state client.AuthorizationState) {
		called1 = true
	}
	mockNotify2 := func(state client.AuthorizationState) {
		called2 = true
	}

	// Подписываем первого подписчика
	service.Subscribe(mockNotify1)
	assert.Equal(t, 1, len(service.subscribers))

	// Подписываем второго подписчика
	service.Subscribe(mockNotify2)
	assert.Equal(t, 2, len(service.subscribers))

	// Проверяем, что функции еще не вызывались
	assert.False(t, called1)
	assert.False(t, called2)
}

func TestAuthService_broadcast(t *testing.T) {
	t.Parallel()

	// Тестирование concurrent code
	synctest.Run(func() {
		// Тестовые состояния авторизации (используемые в коде)
		states := []client.AuthorizationState{
			&client.AuthorizationStateWaitPhoneNumber{},
			&client.AuthorizationStateWaitCode{},
			&client.AuthorizationStateWaitPassword{},
		}

		telegramRepo := mocks.NewTelegramRepo(t)
		service := New(telegramRepo)

		// Тест: broadcast без подписчиков - нормально
		for _, state := range states {
			service.broadcast(state)
		}

		// Создаем несколько подписчиков (реалистичное количество)
		const numSubscribers = 2
		var callCounts [numSubscribers]int64

		// Каналы для получения состояний от каждого подписчика
		statesChans := make([]chan client.AuthorizationState, numSubscribers)
		for i := range statesChans {
			statesChans[i] = make(chan client.AuthorizationState, len(states))
		}

		// Создаем подписчиков
		for i := range numSubscribers {
			mockNotify := func(state client.AuthorizationState) {
				atomic.AddInt64(&callCounts[i], 1)
				statesChans[i] <- state
			}
			service.Subscribe(mockNotify)
		}

		// Отправляем все состояния
		for _, state := range states {
			service.broadcast(state)
		}

		// Ждем завершения горутин
		time.Sleep(100 * time.Millisecond)

		// Проверяем, что каждый подписчик получил все состояния
		for i := range numSubscribers {
			// Проверяем количество вызовов
			assert.Equal(t, int64(len(states)), atomic.LoadInt64(&callCounts[i]),
				"Подписчик %d получил неправильное количество состояний", i)

			// Собираем полученные состояния
			var receivedStates []client.AuthorizationState
			for range len(states) {
				select {
				case state := <-statesChans[i]:
					receivedStates = append(receivedStates, state)
				case <-time.After(1 * time.Second):
					t.Fatalf("Подписчик %d не получил все состояния", i)
				}
			}

			// Проверяем, что все ожидаемые состояния получены
			require.Len(t, receivedStates, len(states))

			expectedStates := make(map[string]bool) // stateType -> found
			for _, state := range states {
				expectedStates[state.AuthorizationStateType()] = false
			}

			// Отмечаем найденные состояния
			for _, receivedState := range receivedStates {
				stateType := receivedState.AuthorizationStateType()
				expectedStates[stateType] = true
			}

			// Проверяем, что все ожидаемые состояния были найдены
			for stateType, found := range expectedStates {
				assert.True(t, found, "Подписчик %d не получил состояние %s", i, stateType)
			}
		}
	})
}
