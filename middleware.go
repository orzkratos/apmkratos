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

func Middleware() middleware.Middleware {
	var m1 middleware.Middleware = NewApmTraceMiddleware() //追踪流程
	var m2 middleware.Middleware = NewRecoveryMiddleware() //追踪崩溃
	return func(handleFunc middleware.Handler) middleware.Handler {
		return m1(m2(handleFunc))
	}
}

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
				opts.TraceContext = obtainIncomingMetadataTraceContext(tsp.RequestHeader())
				apmTx := apm.DefaultTracer().StartTransactionOptions(tsp.Operation(), "request", opts)
				ctx = apm.ContextWithTransaction(ctx, apmTx)
				ctx = newOutgoingContextWithTraceContext(ctx, apmTx.TraceContext())
				defer apmTx.End()
				defer setGRPCContext(&apmTx.Context)
			case transport.KindHTTP:
				htx := GetHttpTspFromContext(ctx)
				req := htx.Request()
				requestName := apmhttp.ServerRequestName(req)
				apmTx, body, req := apmhttp.StartTransactionWithBody(apm.DefaultTracer(), requestName, req)
				ctx = apm.ContextWithTransaction(ctx, apmTx)
				ctx = newOutgoingContextWithTraceContext(ctx, apmTx.TraceContext())
				defer apmTx.End()
				defer setHTTPContext(&apmTx.Context, body, req)
			}
			res, err := handleFunc(ctx, arg)
			if err != nil {
				apmTx := apm.TransactionFromContext(ctx)
				defer apmTx.End()
				e := apm.DefaultTracer().NewError(err)
				e.SetTransaction(apmTx)
				e.Send()
			}
			return res, err
		}
	}
}

func setGRPCContext(ctx *apm.Context) {
	ctx.SetFramework("kratos", "v2?")
	// TODO:
	// doc: https://github.com/elastic/apm-agent-go/blob/main/module/apmgrpc/server.go
}

func setHTTPContext(ctx *apm.Context, body *apm.BodyCapturer, req *nethttp.Request) {
	ctx.SetFramework("kratos", "v2?")
	ctx.SetHTTPRequest(req)
	ctx.SetHTTPRequestBody(body)
	// TODO:
	// ctx.SetHTTPStatusCode(c.Writer.Status())
	// ctx.SetHTTPResponseHeaders(c.Writer.Header())
}

func obtainIncomingMetadataTraceContext(md transport.Header) apm.TraceContext {
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

func GetHttpTspFromContext(ctx context.Context) *http.Transport {
	if txp, ok := transport.FromServerContext(ctx); ok {
		if txp.Kind() == transport.KindHTTP {
			if tsp, ok := txp.(*http.Transport); ok {
				return tsp
			}
		}
	}
	return nil
}

func NewRecoveryMiddleware() middleware.Middleware {
	return recovery.Recovery(NewRecoveryApmOptions(func(format string, args ...interface{}) *errors.Error {
		return errors.Newf(500, "INTERNAL-SERVER-ERROR", format, args...)
	}))
}

func NewRecoveryApmOptions(efn func(format string, args ...interface{}) *errors.Error) recovery.Option {
	return recovery.WithHandler(func(ctx context.Context, req, erx interface{}) error {
		apmTx := apm.TransactionFromContext(ctx)
		if apmTx == nil { // 注意这时 recovery.Recovery() middle 不能设置在 apm middleware 的前面
			return efn("service panic (no apm) erx=%v", erx)
		}
		defer apmTx.End()
		e := apm.DefaultTracer().Recovered(erx)
		e.SetTransaction(apmTx)
		e.Send()
		return efn("service panic erx=%v", erx) //该返回值将被调用层拿到
	})
}
