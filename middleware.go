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
	"context"
	nethttp "net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"
	"google.golang.org/grpc/metadata"
)

// Middleware creates comprehensive APM middleware stack combining request tracing with panic detection
// Implements stacked middleware architecture where trace collection wraps panic protection logic
// Returns optimized middleware chain enabling complete tracing across service request lifecycle
// Handles both standard request flows and panic situations through unified APM reporting
//
// 创建综合 APM 中间件栈，将请求追踪与 panic 检测相结合
// 实现分层中间件架构，追踪收集包装 panic 保护逻辑
// 返回优化的中间件链，实现服务请求生命周期的完整可观测性
// 通过统一的 APM 上报自动处理正常请求流和异常 panic 情况
func Middleware() middleware.Middleware {
	var m1 = NewApmTraceMiddleware() // Trace request flow // 追踪请求流程
	var m2 = NewRecoveryMiddleware() // Trace panic detection // 追踪 panic 检测
	return func(handleFunc middleware.Handler) middleware.Handler {
		return m1(m2(handleFunc))
	}
}

// NewApmTraceMiddleware creates advanced distributed tracing middleware component with gRPC and HTTP support
// Implements intelligent transaction tracking across both gRPC and HTTP transport mechanisms
// Uses W3C TraceContext propagation standards enabling seamless cross-service request correlation
// Captures request metadata and error conditions through comprehensive APM integration
//
// 创建具有多协议支持的高级分布式追踪中间件组件
// 实现跨 gRPC 和 HTTP 传输机制的智能事务追踪
// 利用 W3C TraceContext 传播标准实现无缝跨服务请求关联
// 通过全面的 APM 集成自动捕获请求元数据和错误条件
func NewApmTraceMiddleware() middleware.Middleware {
	return func(handleFunc middleware.Handler) middleware.Handler {
		return func(ctx context.Context, arg interface{}) (interface{}, error) {
			tsp, ok := transport.FromServerContext(ctx)
			if !ok {
				return handleFunc(ctx, arg)
			}

			switch tsp.Kind() {
			case transport.KindGRPC:
				var opts apm.TransactionOptions
				opts.TraceContext = parseIncomingMetadataTraceContext(tsp.RequestHeader())
				apmTx := apm.DefaultTracer().StartTransactionOptions(tsp.Operation(), "request", opts)
				ctx = apm.ContextWithTransaction(ctx, apmTx)
				ctx = newOutgoingContextWithTraceContext(ctx, apmTx.TraceContext())
				defer apmTx.End()
				defer setGRPCContext(&apmTx.Context)
			case transport.KindHTTP:
				httpTransport := GetHttpTransportFromContext(ctx)
				req := httpTransport.Request()
				requestName := apmhttp.ServerRequestName(req)
				apmTx, body, req := apmhttp.StartTransactionWithBody(apm.DefaultTracer(), requestName, req)
				ctx = apm.ContextWithTransaction(ctx, apmTx)
				ctx = newOutgoingContextWithTraceContext(ctx, apmTx.TraceContext())
				defer apmTx.End()
				defer setHTTPContext(&apmTx.Context, body, req)
			}
			res, erk := handleFunc(ctx, arg)
			if erk != nil {
				apmTx := apm.TransactionFromContext(ctx)
				if apmTx != nil {
					defer apmTx.End()
					e := apm.DefaultTracer().NewError(erk)
					e.SetTransaction(apmTx)
					e.Send()
				}
			}
			return res, erk
		}
	}
}

// setGRPCContext configures framework identification metadata within gRPC transaction context
// APM dashboard can show framework-specific insights and optimize gRPC monitoring
//
// 在 gRPC 事务上下文中配置框架识别元数据
// 使 APM 仪表板显示框架特定的洞察并优化 gRPC 监控
func setGRPCContext(ctx *apm.Context) {
	ctx.SetFramework("kratos", "v2?")
	// More gRPC context metadata can be configured here when needed
	// 需要时可以在此配置更多 gRPC 上下文元数据
	// Reference: https://github.com/elastic/apm-agent-go/blob/main/module/apmgrpc/server.go
}

// setHTTPContext configures framework identification and captures comprehensive HTTP request metadata
// Extracts request headers and content enabling deep transaction analysis
//
// 配置框架识别并捕获全面的 HTTP 请求元数据
// 自动提取请求头和正文内容，实现深入的事务分析
func setHTTPContext(ctx *apm.Context, body *apm.BodyCapturer, req *nethttp.Request) {
	ctx.SetFramework("kratos", "v2?")
	ctx.SetHTTPRequest(req)
	ctx.SetHTTPRequestBody(body)
	// Note: Response status code and headers need to be configured following request processing
	// But Kratos middleware design doesn't expose response stream mechanisms
	// Response metadata tracking needs custom implementation when needed
	//
	// 注意：响应状态码和响应头需要在请求处理后配置
	// 但是 Kratos 中间件设计不公开响应流机制
	// 需要时响应元数据追踪需要自定义包装实现
}

// parseIncomingMetadataTraceContext extracts W3C TraceContext propagation data from incoming request metadata
// Implements intelligent parsing with safe degradation returning blank context when extraction fails
// Enables seamless distributed trace correlation across microservice boundaries through standard headers
//
// 从传入请求元数据中提取 W3C TraceContext 传播数据
// 实现智能解析，提取失败时优雅降级返回空上下文
// 通过标准头实现跨微服务边界的无缝分布式追踪关联
func parseIncomingMetadataTraceContext(md transport.Header) apm.TraceContext {
	if value := md.Get(strings.ToLower(apmhttp.W3CTraceparentHeader)); value != "" {
		apmTraceContext, err := apmhttp.ParseTraceparentHeader(value)
		if err != nil {
			return apm.TraceContext{}
		}
		apmTraceContext.State, _ = apmhttp.ParseTracestateHeader(md.Get(strings.ToLower(apmhttp.TracestateHeader)))
		return apmTraceContext
	}
	return apm.TraceContext{}
}

// newOutgoingContextWithTraceContext injects W3C TraceContext propagation headers into outbound request metadata
// Implements automatic trace correlation enabling downstream services to continue transaction tracking
// Maintains trace state alignment across complex microservice chains through standard headers
//
// 将 W3C TraceContext 传播头注入到出站请求元数据中
// 实现自动追踪关联，使下游服务能够继续事务追踪
// 通过标准头注入在复杂微服务调用链中保持追踪状态一致性
func newOutgoingContextWithTraceContext(ctx context.Context, apmTraceContext apm.TraceContext) context.Context {
	apmTraceParentValue := apmhttp.FormatTraceparentHeader(apmTraceContext)
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.Pairs(strings.ToLower(apmhttp.W3CTraceparentHeader), apmTraceParentValue)
	} else {
		md = md.Copy()
		md.Set(strings.ToLower(apmhttp.W3CTraceparentHeader), apmTraceParentValue)
	}
	if apmTraceState := apmTraceContext.State.String(); apmTraceState != "" {
		md.Set(strings.ToLower(apmhttp.TracestateHeader), apmTraceState)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// GetHttpTransportFromContext performs intelligent HTTP transport extraction from request context
// Implements type-safe casting with nil-safe validation ensuring robust transport access patterns
// Returns nil when context lacks HTTP transport enabling safe degradation in mixed transport environments
//
// 从请求上下文中执行智能 HTTP 传输提取
// 实现类型安全转换和 nil 安全验证，确保健壮的传输访问模式
// 当上下文缺少 HTTP 传输时返回 nil，在混合协议环境中实现优雅降级
func GetHttpTransportFromContext(ctx context.Context) *http.Transport {
	if tsp, ok := transport.FromServerContext(ctx); ok {
		if tsp.Kind() == transport.KindHTTP {
			if res, ok := tsp.(*http.Transport); ok {
				return res
			}
		}
	}
	return nil
}

// NewRecoveryMiddleware constructs intelligent panic detection middleware with comprehensive APM integration
// Implements automatic crash capture mechanism transforming runtime panics into structured APM error events
// Returns configured middleware enabling production-grade panic protection with complete tracing
//
// 构造具有全面 APM 集成的智能 panic 检测中间件
// 实现自动崩溃捕获机制，将运行时 panic 转换为结构化 APM 错误事件
// 返回配置的中间件，实现生产级 panic 保护和完整可观测性
func NewRecoveryMiddleware() middleware.Middleware {
	return recovery.Recovery(NewRecoveryApmOptions(func(format string, args ...interface{}) *errors.Error {
		return errors.Newf(500, "INTERNAL-SERVER-ERROR", format, args...)
	}))
}

// NewRecoveryApmOptions constructs panic detection configuration with customizable error transformation logic
// Enables flexible error response generation through provided transformation function efn
// Implements intelligent transaction binding ensuring panic events associate with correct APM transactions
//
// 构造具有可自定义错误转换逻辑的 panic 检测配置
// 通过提供的转换函数 efn 实现灵活的错误响应生成
// 实现智能事务绑定，确保 panic 事件与正确的 APM 事务关联
func NewRecoveryApmOptions(efn func(format string, args ...interface{}) *errors.Error) recovery.Option {
	return recovery.WithHandler(func(ctx context.Context, req, erx interface{}) error {
		apmTx := apm.TransactionFromContext(ctx)
		if apmTx == nil {
			// Important: Place panic detection middleware following APM tracing middleware in chain
			// 重要：在链中将 panic 检测中间件放置在 APM 追踪中间件之后
			return efn("service panic (no apm) erx=%v", erx)
		}
		defer apmTx.End()
		e := apm.DefaultTracer().Recovered(erx)
		e.SetTransaction(apmTx)
		e.Send()
		// Upstream layers process the returned error message
		// 上游层处理返回的错误消息
		return efn("service panic erx=%v", erx)
	})
}
