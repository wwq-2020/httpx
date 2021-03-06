package httpx

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Get Get
func Get(ctx context.Context, url string, resp interface{}, opts ...Option) error {
	return do(ctx, http.MethodGet, url, nil, resp, opts...)
}

// Post Post
func Post(ctx context.Context, url string, req, resp interface{}, opts ...Option) error {
	return do(ctx, http.MethodPost, url, req, resp, opts...)
}

// Put Put
func Put(ctx context.Context, url string, req, resp interface{}, opts ...Option) error {
	return do(ctx, http.MethodPut, url, req, resp, opts...)
}

// Delete Delete
func Delete(ctx context.Context, url string, req, resp interface{}, opts ...Option) error {
	return do(ctx, http.MethodDelete, url, req, resp, opts...)
}

func do(ctx context.Context, method, url string, req, resp interface{}, opts ...Option) error {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	var reqBody io.Reader
	var reqData []byte
	if req != nil {
		var err error
		reqData, err = options.codec.Encode(req)
		if err != nil {
			return errors.WithStack(err)
		}
		reqBody = bytes.NewReader(reqData)
	}
	httpReq, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return errors.WithStack(err)
	}
	httpReq = httpReq.WithContext(ctx)
	if options.reqInterceptor != nil {
		if err := options.reqInterceptor(httpReq); err != nil {
			return errors.WithStack(err)
		}
	}

	httpResp, err := options.client.Do(httpReq)
	if err != nil {
		return errors.WithStack(err)
	}

	respData, respBody, err := drainBody(httpResp.Body)
	if err != nil {
		return errors.WithStack(err)
	}
	defer httpResp.Body.Close()

	httpResp.Body = respBody

	if options.respInterceptor != nil {
		if err := options.respInterceptor(httpResp); err != nil {
			return errors.WithStack(err)
		}
	}
	if resp != nil {
		if err := options.codec.Decode(respData, resp); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func drainBody(src io.ReadCloser) ([]byte, io.ReadCloser, error) {
	defer src.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(src); err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return buf.Bytes(), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

// Client Client
func Client() *http.Client {
	return &http.Client{
		Transport: Transport(),
	}
}

// Transport Transport
func Transport() http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

type retriableTransport struct {
	maxRetry int
	rt       http.RoundTripper
}

func (rt *retriableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var saved []byte
	var err error
	if rt.maxRetry > 0 && req.Body != nil {
		saved, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	var resp *http.Response
	for i := 0; i < rt.maxRetry; i++ {
		req.Body = io.NopCloser(bytes.NewBuffer(saved))
		resp, err = rt.rt.RoundTrip(req)
		if err == nil {
			if resp.StatusCode >= http.StatusInternalServerError {
				time.Sleep(time.Millisecond * 100)
				continue
			}
			return resp, nil
		}
	}
	if resp != nil {
		return resp, nil
	}
	return nil, errors.WithStack(err)
}

// RetriableTransport RetriableTransport
func RetriableTransport(maxRetry int, rt http.RoundTripper) http.RoundTripper {
	return &retriableTransport{
		rt:       rt,
		maxRetry: maxRetry,
	}
}

// RetriableClient RetriableClient
func RetriableClient() *http.Client {
	return &http.Client{
		Transport: RetriableTransport(3, Transport()),
	}
}
