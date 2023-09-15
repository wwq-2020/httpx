package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel"
)

func TestPost(t *testing.T) {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		panic("failed to New")
	}
	bsp := trace.NewBatchSpanProcessor(exp)
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSpanProcessor(bsp),
	)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tp)

	type req struct {
		Data string
	}
	type resp struct {
		Data string
	}
	server := httptest.NewServer(DefaultHandlerWrapper(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq := &req{}
		if err := json.NewDecoder(r.Body).Decode(gotReq); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := json.NewEncoder(w).Encode(&resp{gotReq.Data}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})))
	defer server.Close()
	givenData := "hello world"
	givenReq := &req{
		Data: givenData,
	}

	gotResp := &resp{}
	builder := BaseURL(server.URL)

	if err := builder.Post("").
		WithReq(givenReq).
		WithResp(gotResp).
		Do(context.TODO()); err != nil {
		panic(err)
	}
	if gotResp.Data != givenData {
		t.Fatalf("expected data:%s,got:%s", givenData, gotResp.Data)
	}
}

func TestStatusPost(t *testing.T) {
	type req struct {
		Data string
	}
	type resp struct {
		Data string
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq := &req{}
		if err := json.NewDecoder(r.Body).Decode(gotReq); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := json.NewEncoder(w).Encode(&statusResp{Data: &resp{gotReq.Data}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()
	givenData := "hello world"
	givenReq := &req{
		Data: givenData,
	}

	gotResp := &resp{}

	if err := Post(server.URL).
		WithReq(givenReq).
		WithResp(gotResp).
		WithCodec(&StatusJsonCodec{}).
		Do(context.TODO()); err != nil {
		panic(err)
	}
	if gotResp.Data != givenData {
		t.Fatalf("expected data:%s,got:%s", givenData, gotResp.Data)
	}
}

func Test_builder_BuildHTTPReq(t *testing.T) {
	b := Get("https://www.abcd123.top/api/v1/login?q1=a&q2=b").
		WithQueryString("q3", "c")

	httpReq, err := b.BuildHTTPReq(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("httpReq: %#v", httpReq)

	t.Logf("builder: %#v", b)
}
