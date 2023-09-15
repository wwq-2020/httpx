package httpx

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"log/slog"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultTransprtTimeout = time.Second * 10
)

type TransportFunc func(*http.Request) (*http.Response, error)

func (t TransportFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return t(r)
}

type TransportWrapper func(http.RoundTripper) http.RoundTripper

// TimeoutTransport 添加timeout
func TimeoutTransport(timeout time.Duration) TransportWrapper {
	if timeout <= 0 {
		timeout = defaultTransprtTimeout
	}
	return func(next http.RoundTripper) http.RoundTripper {
		return TransportFunc(func(httpReq *http.Request) (*http.Response, error) {
			ctx, cancel := context.WithTimeout(httpReq.Context(), timeout)
			defer cancel()

			httpReq = httpReq.WithContext(ctx)
			return next.RoundTrip(httpReq)
		})
	}
}

// HeaderTransport 添加header kv
func HeaderTransport(key, value string) TransportWrapper {
	return func(next http.RoundTripper) http.RoundTripper {
		return TransportFunc(func(httpReq *http.Request) (*http.Response, error) {
			httpReq.Header.Add(key, value)
			return next.RoundTrip(httpReq)
		})
	}
}

const (
	ContentTypeKey  = "Content-Type"
	ContentTypeJson = "application/json"
)

// JsonTransport 添加json header
func JsonTransport(next http.RoundTripper) http.RoundTripper {
	return TransportFunc(func(httpReq *http.Request) (*http.Response, error) {
		if httpReq.Header.Get(ContentTypeKey) == "" {
			httpReq.Header.Add(ContentTypeKey, ContentTypeJson)
		}
		return next.RoundTrip(httpReq)
	})
}

// HeadersTransport 添加header
func HeadersTransport(header http.Header) TransportWrapper {
	return func(next http.RoundTripper) http.RoundTripper {
		return TransportFunc(func(httpReq *http.Request) (*http.Response, error) {
			for key, values := range header {
				for _, value := range values {
					httpReq.Header.Add(key, value)
				}
			}
			return next.RoundTrip(httpReq)
		})
	}
}

// LoggingTransport 添加日志
func LoggingTransport(loggingReqBody, loggingRespBody bool) TransportWrapper {
	return func(next http.RoundTripper) http.RoundTripper {
		return TransportFunc(func(httpReq *http.Request) (*http.Response, error) {
			fmt.Println("=======", httpReq.Header)
			spanContext := trace.SpanFromContext(httpReq.Context()).SpanContext()

			traceID := spanContext.TraceID().String()
			spanID := spanContext.SpanID().String()
			kvs := []interface{}{
				"http_method", httpReq.Method,
				"http_url", httpReq.URL.String(),
				"traceID", traceID,
				"spanID", spanID,
			}
			defer func() {
				slog.Info("got http resp", kvs...)
			}()

			isUpgrade := httpReq.Header.Get("Connection") == "Upgrade"
			if !isUpgrade && loggingReqBody && httpReq.Body != nil {
				reqData, reqBody, err := DrainBody(httpReq.Body)
				if err != nil {
					return nil, err
				}
				kvs = append(kvs, "req_data", string(reqData))
				httpReq.Body = reqBody
			}
			slog.Info("send http req", kvs...)
			httpResp, err := next.RoundTrip(httpReq)
			if err != nil {
				return nil, err
			}
			kvs = append(kvs, "http_status_code", httpResp.StatusCode)
			if !isUpgrade && loggingRespBody {
				respData, respBody, err := DrainBody(httpResp.Body)
				if err != nil {
					return nil, err
				}
				kvs = append(kvs, "resp_data", string(respData))
				httpResp.Body = respBody
			}
			return httpResp, nil
		})
	}
}

// TracingTransport 添加traceid
func TracingTransport(serviceName string) TransportWrapper {
	if serviceName == "" {
		serviceName = os.Args[0]
	}
	return func(next http.RoundTripper) http.RoundTripper {
		transport := otelhttp.NewTransport(next, otelhttp.WithServerName(serviceName))
		return TransportFunc(func(httpReq *http.Request) (*http.Response, error) {
			return transport.RoundTrip(httpReq)
		})
	}
}

// StatusCodeTransport 添加statuscode检查
func StatusCodeTransport(expectedStatusCode int) TransportWrapper {
	return func(next http.RoundTripper) http.RoundTripper {
		return TransportFunc(func(httpReq *http.Request) (*http.Response, error) {
			httpResp, err := next.RoundTrip(httpReq)
			if err != nil {
				return nil, err
			}
			if gotStatusCode := httpResp.StatusCode; gotStatusCode != expectedStatusCode {
				return nil, fmt.Errorf("expected statuscode:%d,got:%d", expectedStatusCode, gotStatusCode)
			}
			return httpResp, nil
		})
	}
}

func StatusCodesTransport(expectedStatusCodes ...int) TransportWrapper {
	expectedStatusCodesMap := make(map[int]struct{})
	for _, exexpectedStatusCode := range expectedStatusCodes {
		expectedStatusCodesMap[exexpectedStatusCode] = struct{}{}
	}
	return func(next http.RoundTripper) http.RoundTripper {
		return TransportFunc(func(httpReq *http.Request) (*http.Response, error) {

			httpResp, err := next.RoundTrip(httpReq)
			if err != nil {
				return nil, err
			}
			gotStatusCode := httpResp.StatusCode
			if _, exist := expectedStatusCodesMap[gotStatusCode]; !exist {
				return nil, fmt.Errorf("expected statuscodes:%d,got:%d", expectedStatusCodes, gotStatusCode)
			}
			return httpResp, nil
		})
	}
}

func DefaultTransportWrapper(next http.RoundTripper) TransportFunc {
	for _, wrapper := range []TransportWrapper{
		StatusCodeTransport(http.StatusOK),
		JsonTransport,
		LoggingTransport(true, true),
		TracingTransport(""),
		TimeoutTransport(defaultHandlerTimeout),
	} {
		next = wrapper(next)
	}
	return TransportFunc(func(httpReq *http.Request) (*http.Response, error) {
		resp, err := next.RoundTrip(httpReq)
		if err != nil {
			return nil, err
		}
		return resp, nil

	})
}

func WrapTransport(next http.RoundTripper, wrappers ...TransportWrapper) TransportFunc {
	for _, wrapper := range wrappers {
		next = wrapper(next)
	}
	return TransportFunc(func(httpReq *http.Request) (*http.Response, error) {
		resp, err := next.RoundTrip(httpReq)
		if err != nil {
			return nil, err
		}
		return resp, nil

	})
}
