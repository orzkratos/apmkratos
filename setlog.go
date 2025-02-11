package apmkratos

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/orzkratos/apmkratos/apmkratoslog"
	"go.elastic.co/apm/v2"
)

func SetLogHelper(logger log.Logger, msgKey string) {
	apm.DefaultTracer().SetLogger(apmkratoslog.NewLogHelper(logger, msgKey))
}

func SetApmLogger(logger log.Logger, level log.Level, msgKey string) {
	apm.DefaultTracer().SetLogger(apmkratoslog.NewApmLogger(logger, level, msgKey))
}
