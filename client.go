package httpx

import (
	"net/http"
	"sync"
)

func BuildClient(tws ...TransportWrapper) *http.Client {
	return &http.Client{
		Transport: BuildTransport(tws...),
	}
}

func BuildInsecureClient(tws ...TransportWrapper) *http.Client {
	return &http.Client{
		Transport: BuildInsecureTransport(tws...),
	}
}

func BuildtWrappedClient() *http.Client {
	return &http.Client{
		Transport: BuildWrappedTransport(),
	}
}

func BuildWrappedInsecureClient() *http.Client {
	return &http.Client{
		Transport: BuildWrappedInsecureTransport(),
	}
}

var (
	clientOnce                sync.Once
	client                    *http.Client
	insecureClientOnce        sync.Once
	insecureClient            *http.Client
	wrappedClientOnce         sync.Once
	wrappedClient             *http.Client
	wrappedInsecureClientOnce sync.Once
	wrappedInsecureClient     *http.Client
)

func Client(tws ...TransportWrapper) *http.Client {
	clientOnce.Do(func() {
		client = BuildClient()
	})
	return client
}

func InsecureClient(tws ...TransportWrapper) *http.Client {
	insecureClientOnce.Do(func() {
		insecureClient = BuildInsecureClient()
	})
	return insecureClient
}

func WrappedClient() *http.Client {
	wrappedClientOnce.Do(func() {
		wrappedClient = BuildtWrappedClient()
	})
	return wrappedClient

}

func WrappedInsecureClient() *http.Client {
	wrappedInsecureClientOnce.Do(func() {
		wrappedInsecureClient = BuildWrappedInsecureClient()
	})
	return wrappedInsecureClient

}
