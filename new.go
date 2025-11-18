// Package apmkratos provides advanced Elastic APM middleware integration with Kratos microservice framework
// Implements enterprise-grade distributed tracing capabilities across gRPC and HTTP transport protocols
// Combines intelligent request tracking with panic detection and reporting
// Optimized through W3C TraceContext propagation standards enabling seamless cross-service tracing
//
// apmkratos 包为 Kratos 微服务框架提供高级 Elastic APM 中间件集成
// 实现跨 gRPC 和 HTTP 传输协议的企业级分布式追踪功能
// 结合智能请求追踪与自动 panic 检测和上报机制
// 通过 W3C TraceContext 传播标准优化，实现无缝跨服务追踪
package apmkratos

import (
	"github.com/go-xlan/elasticapm"
	"github.com/go-xlan/elasticapm/apmzaplog"
	"github.com/yyle88/erero"
	"github.com/yyle88/neatjson/neatjsons"
	"github.com/yyle88/zaplog"
	"go.elastic.co/apm/v2"
)

// Initialize bootstraps APM infrastructure using standard environment configuration patterns
// Configures zap-based logging integration and activates default tracing capabilities
// Returns error when initialization fails enabling safe application startup degradation
//
// 使用标准环境配置模式引导 APM 基础设施
// 配置基于 zap 的日志集成并激活默认追踪功能
// 初始化失败时返回错误，使应用程序启动优雅降级
func Initialize(apmConfig *elasticapm.Config) error {
	if err := InitializeWithOptions(apmConfig, elasticapm.NewEnvOption()); err != nil {
		return erero.Wro(err)
	}
	apm.DefaultTracer().SetLogger(apmzaplog.NewLog())
	zaplog.LOG.Debug("Initialize apm success")
	return nil
}

// InitializeWithOptions bootstraps APM infrastructure with advanced customizable environment configuration
// Gives fine-grained setup through explicit environment options and setup hooks
// Supports complex APM deployment scenarios needing non-standard configuration patterns
//
// 使用高级可自定义环境配置引导 APM 基础设施
// 通过显式环境选项规范和可选设置钩子提供精细控制
// 支持需要非标准配置模式的复杂 APM 部署场景
func InitializeWithOptions(apmConfig *elasticapm.Config, envOption *elasticapm.EnvOption, setEnvs ...func()) error {
	zaplog.SUG.Info("Initialize apm apm_config=" + neatjsons.S(apmConfig))
	zaplog.SUG.Info("Initialize apm evo_option=" + neatjsons.S(envOption))

	if err := elasticapm.InitializeWithOptions(apmConfig, envOption, setEnvs...); err != nil {
		return erero.Wro(err)
	}
	zaplog.LOG.Debug("Initialize apm success")
	return nil
}

// Close performs safe APM infrastructure shutdown ensuring complete data transmission
// Must be invoked during application shutdown sequence to prevent trace data loss
// Used with defer pattern in main function to ensure execution regardless of exit path
//
// 执行优雅的 APM 基础设施关闭，确保完整的数据传输
// 必须在应用程序终止序列期间调用，防止追踪数据丢失
// 通常与 main 函数中的 defer 模式一起使用，确保无论退出路径如何都能执行
func Close() {
	elasticapm.Close()
}
