package middleware

import (
	"fmt"

	"github.com/riportdev/riport/share/logger"
)

type RecoveryLogger struct {
	*logger.Logger
}

func NewRecoveryLogger(l *logger.Logger) *RecoveryLogger {
	return &RecoveryLogger{
		Logger: l,
	}
}

func (l *RecoveryLogger) Println(v ...interface{}) {
	l.Errorf(fmt.Sprintln(v...))
}
