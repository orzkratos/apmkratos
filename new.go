package apmkratos

import (
	"github.com/go-xlan/elasticapm"
	"github.com/yyle88/erero"
	"github.com/yyle88/neatjson/neatjsons"
	"github.com/yyle88/zaplog"
)

func INIT(cfg *elasticapm.Config) error {
	var evo = elasticapm.NewEnvOption()

	zaplog.SUG.Info("init apm cfg=" + neatjsons.S(cfg))
	zaplog.SUG.Info("init apm evo=" + neatjsons.S(evo))

	if err := INIT2(cfg, evo); err != nil {
		return erero.Wro(err)
	}

	zaplog.LOG.Debug("init apm success")
	return nil
}

func INIT2(cfg *elasticapm.Config, evo *elasticapm.EnvOption, setEnvs ...func()) error {
	zaplog.SUG.Info("init apm cfg=" + neatjsons.S(cfg))

	if err := elasticapm.INIT2(cfg, evo, setEnvs...); err != nil {
		return erero.Wro(err)
	}

	zaplog.LOG.Debug("init apm success")
	return nil
}

func Close() {
	elasticapm.Close()
}
