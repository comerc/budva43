package util

import (
	"net"
	"time"
)

// IsPortFree проверяет, свободен ли порт
func IsPortFree(addr string) bool {
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
