package apmkratos

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/orzkratos/apmkratos/apmkratos_LOG"
	"go.elastic.co/apm/v2"
)

func SetLog1(logger log.Logger, msgKey string) {
	apm.DefaultTracer().SetLogger(apmkratos_LOG.NewApmLog1Type(logger, msgKey))
}

func SetLog2(logger log.Logger, levels log.Level, msgKey string) {
	apm.DefaultTracer().SetLogger(apmkratos_LOG.NewApmLog2Type(logger, levels, msgKey))
}
