package apmkratoslog

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

/************************
日志实现1
*/

type logHelper struct {
	helper *log.Helper
}

func NewLogHelper(logger log.Logger, msgKey string) *logHelper {
	return &logHelper{
		helper: log.NewHelper(logger, log.WithMessageKey(msgKey)),
	}
}

func (l *logHelper) Debugf(format string, a ...interface{}) {
	l.helper.Debugf(format, a...)
}

func (l *logHelper) Errorf(format string, a ...interface{}) {
	l.helper.Errorf(format, a...)
}

func (l *logHelper) Warningf(format string, a ...interface{}) {
	l.helper.Warnf(format, a...)
}

/************************
日志实现2
*/

// 目前还是不能打印出行号
type apmLogger struct {
	logger log.Logger
	level  log.Level
	msgKey string
}

func NewApmLogger(logger log.Logger, level log.Level, msgKey string) *apmLogger {
	return &apmLogger{
		logger: logger,
		level:  level,
		msgKey: msgKey,
	}
}

func (l *apmLogger) Debugf(format string, a ...interface{}) {
	if log.LevelDebug >= l.level {
		_ = l.logger.Log(log.LevelDebug, l.msgKey, fmt.Sprintf(format, a...))
	}
}

func (l *apmLogger) Errorf(format string, a ...interface{}) {
	if log.LevelError >= l.level {
		_ = l.logger.Log(log.LevelError, l.msgKey, fmt.Sprintf(format, a...))
	}
}

func (l *apmLogger) Warningf(format string, a ...interface{}) {
	if log.LevelWarn >= l.level {
		_ = l.logger.Log(log.LevelWarn, l.msgKey, fmt.Sprintf(format, a...))
	}
}
