package apmkratos

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/orzkratos/apmkratos/apmkratoslog"
	"go.elastic.co/apm/v2"
)

func SetLog(logger log.Logger, levels log.Level) {
	SetLog2(logger, levels, "KRATOS-APM-LOG")
}

func SetLog2(logger log.Logger, levels log.Level, msgKey string) {
	SetLog4(logger, levels, msgKey)
}

func SetLog3(logger log.Logger, msgKey string) {
	apm.DefaultTracer().SetLogger(apmkratoslog.NewApmLog1Type(logger, msgKey))
}

func SetLog4(logger log.Logger, levels log.Level, msgKey string) {
	apm.DefaultTracer().SetLogger(apmkratoslog.NewApmLog2Type(logger, levels, msgKey))
}
