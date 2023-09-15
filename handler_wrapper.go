package httpx

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultHandlerTimeout = time.Second * 10
)

func JsonHandler[Req, Resp any](handler func(ctx context.Context, req Req) (Resp, error)) http.Handler {
	return Handler(defaultCodec, handler)
}

func Handler[Req, Resp any](codec Codec, handler func(ctx context.Context, req Req) (Resp, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqObj := new(Req)
		if err := codec.Decode(r.Body, reqObj); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		respObj, err := handler(ctx, *reqObj)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		respData, err := codec.Encode(respObj)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(respData); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

// TimeoutHandler 添加timeout
func TimeoutHandler(timeout time.Duration) HandlerWrapper {
	if timeout <= 0 {
		timeout = defaultHandlerTimeout
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, httpReq *http.Request) {
			ctx, cancel := context.WithTimeout(httpReq.Context(), timeout)
			defer cancel()
			httpReq = httpReq.WithContext(ctx)
			next.ServeHTTP(w, httpReq)
		})
	}
}

// TracingHandler 添加traceid
func TracingHandler(serviceName string) HandlerWrapper {
	return func(next http.Handler) http.Handler {
		handler := otelhttp.NewHandler(next, "serve http req")
		return http.HandlerFunc(func(w http.ResponseWriter, httpReq *http.Request) {
			handler.ServeHTTP(w, httpReq)
		})
	}
}

// LoggingHandler 添加日志
func LoggingHandler(loggingReqBody, loggingRespBody bool) HandlerWrapper {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, httpReq *http.Request) {
			spanContext := trace.SpanFromContext(httpReq.Context()).SpanContext()

			traceID := spanContext.TraceID().String()
			spanID := spanContext.SpanID().String()
			kvs := []interface{}{
				"http_method", httpReq.Method,
				"http_url", httpReq.URL.String(),
				"traceID", traceID,
				"spanID", spanID,
			}

			isUpgrade := httpReq.Header.Get("Connection") == "Upgrade"
			if !isUpgrade {
				wWrapped := wrapResponseWriter(w)
				if loggingReqBody && httpReq.Body != nil {
					reqData, reqBody, err := DrainBody(httpReq.Body)
					if err != nil {
						return
					}
					httpReq.Body = reqBody

					kvs = append(kvs, "req_data", string(reqData))
				}
				defer func() {
					if loggingRespBody {
						respData := wWrapped.Body()
						statusCode := wWrapped.StatusCode()
						kvs = append(kvs, "resp_data", string(respData), "statusCode", statusCode)
					}
					slog.Info("serve http req", kvs...)
				}()
				next.ServeHTTP(wWrapped, httpReq)
				return
			}

			defer func() {
				slog.Info("serve http req", kvs...)
			}()

			next.ServeHTTP(w, httpReq)
		})
	}
}

// WrappedResponseWriter WrappedResponseWriter
type WrappedResponseWriter interface {
	http.ResponseWriter
	Body() string
	StatusCode() int
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	buf        *bytes.Buffer
}

func (rw *responseWriterWrapper) Flush() {
	rw.ResponseWriter.(http.Flusher).Flush()
}

func (rw *responseWriterWrapper) CloseNotify() <-chan bool {
	return rw.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (rw *responseWriterWrapper) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.statusCode = statusCode
}

func (rw *responseWriterWrapper) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	if err != nil {
		return 0, err
	}
	rw.buf.Write(data[:n])
	return n, nil
}

func (rw *responseWriterWrapper) Body() string {
	body := rw.buf.String()
	rw.buf.Reset()
	return body
}

func (rw *responseWriterWrapper) StatusCode() int {
	if rw.statusCode == 0 {
		return http.StatusOK
	}
	return rw.statusCode
}

func wrapResponseWriter(w http.ResponseWriter) WrappedResponseWriter {
	raw, ok := w.(WrappedResponseWriter)
	if ok {
		return raw
	}
	return &responseWriterWrapper{
		ResponseWriter: w,
		buf:            bytes.NewBuffer(nil),
		statusCode:     http.StatusOK,
	}
}

type HandlerWrapper func(http.Handler) http.Handler

func DefaultHandlerWrapper(next http.Handler) http.Handler {
	for _, wrapper := range []HandlerWrapper{
		LoggingHandler(true, true),
		TracingHandler(""),
		TimeoutHandler(defaultHandlerTimeout),
	} {
		next = wrapper(next)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, httpReq *http.Request) {
		next.ServeHTTP(w, httpReq)
	})
}

func WrapHandler(next http.Handler, wrappers ...HandlerWrapper) http.Handler {
	for _, wrapper := range wrappers {
		next = wrapper(next)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, httpReq *http.Request) {
		next.ServeHTTP(w, httpReq)
	})
}
