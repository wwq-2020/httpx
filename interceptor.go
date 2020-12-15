package httpx

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ReqInterceptor ReqInterceptor
type ReqInterceptor func(*http.Request) error

// RespInterceptor RespInterceptor
type RespInterceptor func(*http.Response) error

// ContentTypeReqInterceptor ContentTypeReqInterceptor
func ContentTypeReqInterceptor(contentType string) ReqInterceptor {
	return func(httpReq *http.Request) error {
		httpReq.Header.Set(ContentTypeHeader, contentType)
		return nil
	}
}

// LoggingReqInterceptor LoggingReqInterceptor
func LoggingReqInterceptor(httpReq *http.Request) error {
	logger := logrus.WithField("method", httpReq.Method).
		WithField("url", httpReq.URL.String())
	if httpReq.Body != nil {
		reqData, reqBody, err := drainBody(httpReq.Body)
		if err != nil {
			return errors.WithStack(err)
		}
		httpReq.Body = reqBody
		logger = logger.WithField("reqData", string(reqData))
	}
	logger.Info("do http req")
	return nil
}

// ChainedReqInterceptor ChainedReqInterceptor
func ChainedReqInterceptor(reqInterceptors ...ReqInterceptor) ReqInterceptor {
	return func(httpReq *http.Request) error {
		for _, reqInterceptor := range reqInterceptors {
			if err := reqInterceptor(httpReq); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}
}

// StatusCodeRespInterceptor StatusCodeRespInterceptor
func StatusCodeRespInterceptor(expected int) RespInterceptor {
	return func(httpResp *http.Response) error {
		got := httpResp.StatusCode
		if got == expected {
			return nil
		}
		return fmt.Errorf("expected statuscode:%d, got:%d", expected, got)
	}
}

// StatusCodeRangeRespInterceptor StatusCodeRangeRespInterceptor
func StatusCodeRangeRespInterceptor(codeStart, codeEnd int) RespInterceptor {
	return func(httpResp *http.Response) error {
		got := httpResp.StatusCode
		if got < codeStart || got > codeEnd {
			return fmt.Errorf("expected code in (%d,%d), got:%d", codeStart, codeEnd, got)
		}
		return nil
	}
}

// LoggingRespInterceptor LoggingRespInterceptor
func LoggingRespInterceptor(httpResp *http.Response) error {
	logger := logrus.WithField("statuscode", httpResp.StatusCode)
	if httpResp.Body != nil {
		respData, respBody, err := drainBody(httpResp.Body)
		if err != nil {
			return errors.WithStack(err)
		}
		logger = logger.WithField("respData", string(respData))
		httpResp.Body = respBody
	}

	logger.Info("got http resp")
	return nil
}

// ChainedRespInterceptor ChainedRespInterceptor
func ChainedRespInterceptor(respInterceptors ...RespInterceptor) RespInterceptor {
	return func(httpResp *http.Response) error {
		for _, respInterceptor := range respInterceptors {
			if err := respInterceptor(httpResp); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}
}
