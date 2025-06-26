package log

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
)

var loggerNamesInstance *loggerNames

type loggerNames struct {
	mu sync.RWMutex
	m  map[uint64]string // GID -> loggerName
}

func createLoggerNames() {
	loggerNamesInstance = &loggerNames{
		m: make(map[uint64]string),
	}
}

// getGID возвращает ID текущей горутины
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

// SetLoggerName устанавливает loggerName для текущей горутины
func SetLoggerName(loggerName string) {
	loggerNamesInstance.mu.Lock()
	defer loggerNamesInstance.mu.Unlock()
	gid := getGID()
	loggerNamesInstance.m[gid] = loggerName
}

// GetLoggerName возвращает loggerName для текущей горутины
func GetLoggerName() string {
	loggerNamesInstance.mu.RLock()
	defer loggerNamesInstance.mu.RUnlock()
	gid := getGID()
	return loggerNamesInstance.m[gid]
}

func GetPackageFileNameWithLine() string {
	callInfo := GetCallStack(3, false)[0]
	return fmt.Sprintf("%s:%d", callInfo.FileName, callInfo.Line)
}

// NewLogger автоматически определяет loggerName
func NewLogger() *Logger {
	var loggerName string
	testing := os.Getenv("GOEXPERIMENT") == "synctest"
	if testing {
		loggerName = GetLoggerName()
	} else {
		loggerName = GetPackageFileNameWithLine()
	}
	return newLogger(loggerName)
}
