package apmkratos

import (
	"github.com/go-xlan/elasticapm"
	"github.com/yyle88/zaplog"
	"go.elastic.co/apm/v2"
	"go.uber.org/zap"
)

// CheckApmAgentVersion 检查apm包版本是否相同，这是因为apm作为单独的模块，假如使用的版本不同，逻辑就无法正常执行
// 因此建议是所有使用到 apm 的三方包都能够实现这个函数
// 以确保不同模块使用的apm包版本相同
func CheckApmAgentVersion(version string) bool {
	if agentVersion := apm.AgentVersion; version != agentVersion {
		zaplog.LOGGER.LOG.Warn("check apm agent versions not match", zap.String("arg_version", version), zap.String("pkg_version", agentVersion))
		return false
	}
	return elasticapm.CheckApmAgentVersion(version) //把依赖的模块也检查检查
}

// GetApmAgentVersion 获得版本号
// 假如你用的我的包，我的包 go.mod 里面引用的 apm 是 v2.0.0 的，这里就会返回 v2.0.0 版本号字符串
// 假如你项目里直接用到 apm 包，则你需要检查你的包和我的包版本是否相同
// 假如不同，逻辑不通
func GetApmAgentVersion() string {
	return apm.AgentVersion
}
