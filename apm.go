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
	"github.com/yyle88/zaplog"
	"go.elastic.co/apm/v2"
	"go.uber.org/zap"
)

// GetApmAgentVersion retrieves the Elastic APM agent version ID embedded within this package
// Returns version string enabling APM agent version checks at application scope
// Important when coordinating packages that depend on specific APM agent releases
//
// 检索嵌入在此包中的 Elastic APM agent 版本标识符
// 返回版本字符串，使应用程序级别能够验证 APM agent 兼容性
// 在协调依赖特定 APM agent 发布的多个包时至关重要
func GetApmAgentVersion() string {
	return apm.AgentVersion
}

// CheckApmAgentVersion validates version alignment between given version and package-embedded APM agent
// Performs version checks across the whole package chain including nested modules
// Returns false when version mismatch is detected to prevent runtime conflicts
//
// 验证提供的版本与包嵌入的 APM agent 之间的版本对齐
// 实现跨完整依赖链（包括嵌套模块）的全面兼容性验证
// 检测到版本不匹配时返回 false，防止潜在的运行时不兼容性
func CheckApmAgentVersion(version string) bool {
	if agentVersion := apm.AgentVersion; version != agentVersion {
		zaplog.LOGGER.LOG.Warn("check apm agent versions not match", zap.String("arg_version", version), zap.String("pkg_version", agentVersion))
		return false
	}
	// Check version alignment across the package chain
	// 递归验证依赖模块链中的版本一致性
	return elasticapm.CheckApmAgentVersion(version)
}
