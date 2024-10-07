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
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, arg interface{}) (interface{}, error) {
			tp, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, arg)
			}

			switch tp.Kind() {
			case transport.KindGRPC:
				var opts apm.TransactionOptions
				opts.TraceContext = obtainIncomingMetadataTraceContext(tp.RequestHeader())
				tx := apm.DefaultTracer().StartTransactionOptions(tp.Operation(), "request", opts)
				ctx = apm.ContextWithTransaction(ctx, tx)
				ctx = newOutgoingContextWithTraceContext(ctx, tx.TraceContext())
				defer tx.End()
				defer setGRPCContext(&tx.Context)
			case transport.KindHTTP:
				htx := GetHttpTspFromContext(ctx)
				req := htx.Request()
				requestName := apmhttp.ServerRequestName(req)
				tx, body, req := apmhttp.StartTransactionWithBody(apm.DefaultTracer(), requestName, req)
				ctx = apm.ContextWithTransaction(ctx, tx)
				ctx = newOutgoingContextWithTraceContext(ctx, tx.TraceContext())
				defer tx.End()
				defer setHTTPContext(&tx.Context, body, req)
			}
			res, err := handler(ctx, arg)
			if err != nil {
				tx := apm.TransactionFromContext(ctx)
				defer tx.End()
				e := apm.DefaultTracer().NewError(err)
				e.SetTransaction(tx)
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
	return recovery.Recovery(NewRecoveryOption())
}

func NewRecoveryOption() recovery.Option {
	return NewRecoveryOptionWithErkFunc(func(format string, args ...interface{}) *errors.Error {
		return errors.Newf(500, "INTERNAL-SERVER-ERROR", format, args...)
	})
}

func NewRecoveryOptionWithErkFunc(newErkFunc func(format string, args ...interface{}) *errors.Error) recovery.Option {
	return recovery.WithHandler(func(ctx context.Context, req, err interface{}) error {
		tx := apm.TransactionFromContext(ctx)
		if tx == nil { // 注意这时 recovery.Recovery() middle 不能设置在 apm middleware 的前面
			// 否则将会从这里返回
			return newErkFunc("service panic (no apm) error=%v", err)
		}
		defer tx.End()
		e := apm.DefaultTracer().Recovered(err)
		e.SetTransaction(tx)
		e.Send()
		return newErkFunc("service panic error=%v", err) //该返回值将被调用层拿到
	})
}
