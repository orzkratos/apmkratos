package apmkratoslog

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

/************************
日志实现1
*/

type apmLog1Type log.Helper

func NewApmLog1Type(logger log.Logger, msgKey string) *apmLog1Type {
	return (*apmLog1Type)(log.NewHelper(logger, log.WithMessageKey(msgKey)))
}

func (H *apmLog1Type) Debugf(format string, a ...interface{}) {
	(*log.Helper)(H).Debugf(format, a...)
}

func (H *apmLog1Type) Errorf(format string, a ...interface{}) {
	(*log.Helper)(H).Errorf(format, a...)
}

func (H *apmLog1Type) Warningf(format string, a ...interface{}) {
	(*log.Helper)(H).Warnf(format, a...)
}

/************************
日志实现2
*/

// 目前还是不能打印出行号
type apmLog2Type struct {
	logger log.Logger
	levels log.Level
	msgKey string
}

func NewApmLog2Type(logger log.Logger, levels log.Level, msgKey string) *apmLog2Type {
	return &apmLog2Type{logger: logger, levels: levels, msgKey: msgKey}
}

func (H *apmLog2Type) Debugf(format string, a ...interface{}) {
	const levels = log.LevelDebug
	if levels >= H.levels {
		_ = H.logger.Log(levels, H.msgKey, fmt.Sprintf(format, a...))
	}
}

func (H *apmLog2Type) Errorf(format string, a ...interface{}) {
	const levels = log.LevelError
	if levels >= H.levels {
		_ = H.logger.Log(levels, H.msgKey, fmt.Sprintf(format, a...))
	}
}

func (H *apmLog2Type) Warningf(format string, a ...interface{}) {
	const levels = log.LevelWarn
	if levels >= H.levels {
		_ = H.logger.Log(levels, H.msgKey, fmt.Sprintf(format, a...))
	}
}
