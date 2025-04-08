package user

import (
	"github.com/zelenin/go-tdlib/client"
)

// User представляет расширение структуры User из TDLib
type User struct {
	// Встраиваем структуру из TDLib
	*client.User
}

// NewUser создает новый экземпляр User из client.User
func NewUser(tdlibUser *client.User) *User {
	if tdlibUser == nil {
		return nil
	}

	return &User{
		User: tdlibUser,
	}
}
