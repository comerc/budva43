package storage

import (
	"fmt"

	"github.com/comerc/budva43/app/log"
)

type LoggerAdapter struct {
	log *log.Logger
}

func NewLoggerAdapter(log *log.Logger) *LoggerAdapter {
	return &LoggerAdapter{
		log: log,
	}
}

func (l *LoggerAdapter) Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.log.Error(msg)
}

func (l *LoggerAdapter) Warningf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.log.Warn(msg)
}

func (l *LoggerAdapter) Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.log.Info(msg)
}

func (l *LoggerAdapter) Debugf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.log.Debug(msg)
}
