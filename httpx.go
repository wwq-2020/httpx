package httpx

import (
	"bytes"
	"context"
	"io"
	"net/http"
	stdurl "net/url"
	"time"

	"github.com/google/go-querystring/query"
)

type Builder interface {
	Get(path string) Builder
	Method(method, path string) Builder
	BaseURL(baseURL string) Builder
	Head(path string) Builder
	Post(path string) Builder
	Put(path string) Builder
	Patch(path string) Builder
	Delete(path string) Builder
	Connect(path string) Builder
	Options(path string) Builder
	Trace(path string) Builder
	WithQueryString(key string, value string) Builder
	WithURLValues(values stdurl.Values) Builder
	WithQueryStringObj(obj interface{}) Builder
	WithCodec(codec Codec) Builder
	WithHeader(key string, value string) Builder
	WithBasicAuth(username, password string) Builder
	WithHeaders(headers http.Header) Builder
	WithReq(req interface{}) Builder
	WithResp(resp interface{}) Builder
	ExpectedStatusCodes(...int) Builder
	Logging(loggingReq, loggingResp bool) Builder
	Timeout(timeout time.Duration) Builder
	Tracing(tracing bool) Builder
	ContentType(contentType string) Builder
	Insecure(insecure bool) Builder
	BuildHTTPReq(context.Context) (*http.Request, error)
	BuildTransport(context.Context) (http.RoundTripper, error)
	Do(context.Context) error
	WithTransport(transport http.RoundTripper) Builder
	DoWithTransport(ctx context.Context, transport http.RoundTripper) error
	DoWithClient(ctx context.Context, client *http.Client) error
}

type builder struct {
	path                string
	method              string
	baseURL             string
	codec               Codec
	resp                interface{}
	req                 interface{}
	urlValues           stdurl.Values
	header              http.Header
	expectedStatusCodes []int
	loggingReq          bool
	loggingResp         bool
	timeout             time.Duration
	tracing             bool
	contentType         string
	insecure            bool
	transport           http.RoundTripper
	err                 error
}

func New() Builder {
	return &builder{
		urlValues:           make(stdurl.Values),
		header:              make(http.Header),
		codec:               defaultCodec,
		expectedStatusCodes: []int{http.StatusOK},
		loggingReq:          true,
		loggingResp:         true,
		timeout:             defaultTransprtTimeout,
		tracing:             true,
		contentType:         ContentTypeJson,
		insecure:            false,
	}
}

func Get(path string) Builder {
	return New().Get(path)
}

func Method(method, path string) Builder {
	return New().Method(method, path)
}

func BaseURL(baseURL string) Builder {
	return New().BaseURL(baseURL)
}

func Head(path string) Builder {
	return New().Head(path)
}

func Post(path string) Builder {
	return New().Post(path)
}

func Put(path string) Builder {
	return New().Put(path)
}

func Patch(url string) Builder {
	return New().Patch(url)
}

func Delete(path string) Builder {
	return New().Delete(path)
}

func Connect(path string) Builder {
	return New().Connect(path)
}

func Options(path string) Builder {
	return New().Options(path)
}

func Trace(path string) Builder {
	return New().Trace(path)
}

func WithQueryString(key string, value string) Builder {
	return New().WithQueryString(key, value)
}

func WithQueryStringObj(obj interface{}) Builder {
	return New().WithQueryStringObj(obj)
}

func WithCodec(codec Codec) Builder {
	return New().WithCodec(codec)
}

func WithHeader(key, value string) Builder {
	return New().WithHeader(key, value)
}

func WithBasicAuth(username, password string) Builder {
	return New().WithBasicAuth(username, password)
}

func WithHeaders(headers http.Header) Builder {
	return New().WithHeaders(headers)
}

func WithReq(req interface{}) Builder {
	return New().WithReq(req)
}

func WithResp(resp interface{}) Builder {
	return New().WithResp(resp)
}

func ExpectedStatusCodes(expectedStatusCodes ...int) Builder {
	return New().ExpectedStatusCodes(expectedStatusCodes...)
}
func Logging(loggingReq, loggingResp bool) Builder {
	return New().Logging(loggingReq, loggingResp)
}

func Timeout(timeout time.Duration) Builder {
	return New().Timeout(timeout)
}
func Tracing(tracing bool) Builder {
	return New().Tracing(tracing)
}
func ContentType(contentType string) Builder {
	return New().ContentType(contentType)
}
func Insecure(insecure bool) Builder {
	return New().Insecure(insecure)
}

func WithTransport(transport http.RoundTripper) Builder {
	return New().WithTransport(transport)
}

func (b *builder) Get(path string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.method = http.MethodGet
	newBuilder.path = path
	return newBuilder
}

func (b *builder) Method(method, path string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.method = method
	newBuilder.path = path
	return newBuilder
}

func (b *builder) BaseURL(baseURL string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.baseURL = baseURL
	return newBuilder
}

func (b *builder) Head(path string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.method = http.MethodHead
	newBuilder.path = path
	return newBuilder
}

func (b *builder) Post(path string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.method = http.MethodPost
	newBuilder.path = path
	return newBuilder
}

func (b *builder) Put(path string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.method = http.MethodPut
	newBuilder.path = path
	return newBuilder
}

func (b *builder) Patch(path string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.method = http.MethodPatch
	newBuilder.path = path
	return newBuilder
}

func (b *builder) Delete(path string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.method = http.MethodDelete
	newBuilder.path = path
	return newBuilder
}

func (b *builder) Connect(path string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.method = http.MethodConnect
	newBuilder.path = path
	return newBuilder
}

func (b *builder) Options(path string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.method = http.MethodOptions
	newBuilder.path = path
	return newBuilder
}

func (b *builder) Trace(path string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.method = http.MethodTrace
	newBuilder.path = path
	return newBuilder
}

func (b *builder) WithQueryString(key string, value string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.urlValues.Add(key, value)
	return newBuilder
}

func (b *builder) WithQueryStringObj(obj interface{}) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	urlValues, err := query.Values(obj)
	newBuilder.err = err
	for key, values := range urlValues {
		for _, value := range values {
			newBuilder.urlValues.Add(key, value)
		}
	}
	return newBuilder
}

func (b *builder) WithURLValues(urlValues stdurl.Values) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	for key, values := range urlValues {
		for _, value := range values {
			newBuilder.urlValues.Add(key, value)
		}
	}
	return newBuilder
}

func (b *builder) WithCodec(codec Codec) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.codec = codec
	return newBuilder
}

func (b *builder) WithHeader(key string, value string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.header.Add(key, value)
	return newBuilder
}

func (b *builder) WithBasicAuth(username, password string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.header.Set("Authorization", "Basic "+BasicAuth(username, password))
	return newBuilder
}

func (b *builder) WithHeaders(headers http.Header) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	for k, vs := range headers {
		for _, v := range vs {
			newBuilder.header.Add(k, v)
		}
	}
	return newBuilder
}

func (b *builder) WithReq(req interface{}) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.req = req
	return newBuilder
}

func (b *builder) WithResp(resp interface{}) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.resp = resp
	return newBuilder
}

func (b *builder) ExpectedStatusCodes(expectedStatusCodes ...int) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.expectedStatusCodes = expectedStatusCodes
	return newBuilder
}

func (b *builder) Logging(loggingReq, loggingResp bool) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.loggingReq = loggingReq
	newBuilder.loggingResp = loggingResp
	return newBuilder
}

func (b *builder) Timeout(timeout time.Duration) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.timeout = timeout
	return newBuilder
}
func (b *builder) Tracing(tracing bool) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.tracing = tracing
	return newBuilder
}
func (b *builder) ContentType(contentType string) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.contentType = contentType
	return newBuilder
}

func (b *builder) Insecure(insecure bool) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.insecure = insecure
	return newBuilder
}

func (b *builder) BuildHTTPReq(ctx context.Context) (*http.Request, error) {
	if b.err != nil {
		return nil, b.err
	}
	url := b.baseURL + b.path
	urlValues := make(stdurl.Values)

	{
		urlObj, err := stdurl.Parse(url)
		if err != nil {
			return nil, err
		}
		for key, values := range b.urlValues {
			for _, value := range values {
				urlValues.Add(key, value)
			}
		}
		for key, values := range urlObj.Query() {
			for _, value := range values {
				urlValues.Add(key, value)
			}
		}
		if len(urlValues) != 0 {
			urlObj.RawQuery = urlValues.Encode()
		}
		url = urlObj.String()
	}

	method := b.method
	if b.method == "" {
		method = http.MethodGet
	}

	var body io.Reader
	var data []byte
	if b.req != nil {
		var err error
		data, err = b.codec.Encode(b.req)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	headers := make(http.Header)
	for key, values := range b.header {
		for _, value := range values {
			headers.Set(key, value)
		}
	}
	contentType := b.contentType
	if contentType == "" {
		contentType = ContentTypeJson
	}
	if headers.Get(ContentTypeKey) == "" {
		headers.Set(ContentTypeKey, contentType)
	}
	httpReq.Header = headers
	return httpReq, nil
}

func (b *builder) BuildTransport(ctx context.Context) (http.RoundTripper, error) {
	if b.err != nil {
		return nil, b.err
	}
	transport := Transport()

	if b.insecure {
		transport = InsecureTransport()
	}
	if b.transport != nil {
		transport = b.transport
	}
	expectedStatusCodes := []int{http.StatusOK}

	if len(b.expectedStatusCodes) != 0 {
		expectedStatusCodes = b.expectedStatusCodes
	}
	tws := []TransportWrapper{
		StatusCodesTransport(expectedStatusCodes...),
	}
	if b.contentType == "" || b.contentType == ContentTypeJson {
		tws = append(tws, JsonTransport)
	}
	tws = append(tws, LoggingTransport(b.loggingReq, b.loggingResp))
	if b.tracing {
		tws = append(tws, TracingTransport(""))
	}
	tws = append(tws, TimeoutTransport(b.timeout))
	transport = WrapTransport(transport, tws...)
	return transport, nil
}

func (b *builder) WithTransport(transport http.RoundTripper) Builder {
	newBuilder := b.clone()
	if newBuilder.err != nil {
		return newBuilder
	}
	newBuilder.transport = transport
	return newBuilder
}

func (b *builder) Do(ctx context.Context) error {
	if b.err != nil {
		return b.err
	}
	transport, err := b.BuildTransport(ctx)
	if err != nil {
		return err
	}
	return b.DoWithTransport(ctx, transport)
}

func (b *builder) DoWithTransport(ctx context.Context, transport http.RoundTripper) error {
	if b.err != nil {
		return b.err
	}
	client := &http.Client{
		Transport: transport,
	}
	return b.DoWithClient(ctx, client)

}

func (b *builder) DoWithClient(ctx context.Context, client *http.Client) error {
	if b.err != nil {
		return b.err
	}
	httpReq, err := b.BuildHTTPReq(ctx)
	if err != nil {
		return err
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	if b.resp != nil {
		err := b.codec.Decode(httpResp.Body, b.resp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *builder) clone() *builder {
	urlValues := make(stdurl.Values)
	for key, values := range b.urlValues {
		for _, value := range values {
			urlValues.Add(key, value)

		}
	}
	header := make(http.Header)
	for key, values := range b.header {
		for _, value := range values {
			header.Add(key, value)

		}
	}
	return &builder{
		path:                b.path,
		method:              b.method,
		baseURL:             b.baseURL,
		codec:               b.codec,
		resp:                b.resp,
		req:                 b.req,
		urlValues:           urlValues,
		header:              header,
		expectedStatusCodes: b.expectedStatusCodes,
		loggingReq:          b.loggingReq,
		loggingResp:         b.loggingResp,
		timeout:             b.timeout,
		tracing:             b.tracing,
		contentType:         b.contentType,
		insecure:            b.insecure,
		err:                 b.err,
		transport:           b.transport,
	}
}
