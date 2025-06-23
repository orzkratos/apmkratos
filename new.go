package apmkratos

import (
	"github.com/go-xlan/elasticapm"
	"github.com/go-xlan/elasticapm/apmzaplog"
	"github.com/yyle88/erero"
	"github.com/yyle88/neatjson/neatjsons"
	"github.com/yyle88/zaplog"
	"go.elastic.co/apm/v2"
)

func Initialize(apmConfig *elasticapm.Config) error {
	if err := InitializeWithOptions(apmConfig, elasticapm.NewEnvOption()); err != nil {
		return erero.Wro(err)
	}
	apm.DefaultTracer().SetLogger(apmzaplog.NewLog())
	zaplog.LOG.Debug("Initialize apm success")
	return nil
}

func InitializeWithOptions(apmConfig *elasticapm.Config, envOption *elasticapm.EnvOption, setEnvs ...func()) error {
	zaplog.SUG.Info("Initialize apm apm_config=" + neatjsons.S(apmConfig))
	zaplog.SUG.Info("Initialize apm evo_option=" + neatjsons.S(envOption))

	if err := elasticapm.InitializeWithOptions(apmConfig, envOption, setEnvs...); err != nil {
		return erero.Wro(err)
	}
	zaplog.LOG.Debug("Initialize apm success")
	return nil
}

func Close() {
	elasticapm.Close()
}
