package httpx

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

func BuildTransport(tws ...TransportWrapper) http.RoundTripper {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	transport := &http.Transport{
		IdleConnTimeout:     30 * time.Second,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     10000,
		MaxIdleConns:        10000,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		DisableCompression:     false,
		DisableKeepAlives:      false,
		ResponseHeaderTimeout:  360 * time.Second,
		ExpectContinueTimeout:  360 * time.Second,
		MaxResponseHeaderBytes: 1 << 20,
		WriteBufferSize:        1 << 12,
		ReadBufferSize:         1 << 12,
		ForceAttemptHTTP2:      false,
	}
	return WrapTransport(transport, tws...)
}

func BuildWrappedTransport() http.RoundTripper {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	transport := &http.Transport{
		IdleConnTimeout:     30 * time.Second,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     1000,
		MaxIdleConns:        1000,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		DisableCompression:     false,
		DisableKeepAlives:      false,
		ResponseHeaderTimeout:  360 * time.Second,
		ExpectContinueTimeout:  360 * time.Second,
		MaxResponseHeaderBytes: 1 << 10,
		WriteBufferSize:        1 << 12,
		ReadBufferSize:         1 << 12,
		ForceAttemptHTTP2:      false,
	}
	return DefaultTransportWrapper(transport)
}

func BuildInsecureTransport(tws ...TransportWrapper) http.RoundTripper {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	transport := &http.Transport{
		IdleConnTimeout:     30 * time.Second,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     1000,
		MaxIdleConns:        1000,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		DisableCompression:     false,
		DisableKeepAlives:      false,
		ResponseHeaderTimeout:  360 * time.Second,
		ExpectContinueTimeout:  360 * time.Second,
		MaxResponseHeaderBytes: 1 << 10,
		WriteBufferSize:        1 << 12,
		ReadBufferSize:         1 << 12,
		ForceAttemptHTTP2:      false,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return WrapTransport(transport, tws...)

}

func BuildWrappedInsecureTransport() http.RoundTripper {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	transport := &http.Transport{
		IdleConnTimeout:     30 * time.Second,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     1000,
		MaxIdleConns:        1000,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		DisableCompression:     false,
		DisableKeepAlives:      false,
		ResponseHeaderTimeout:  360 * time.Second,
		ExpectContinueTimeout:  360 * time.Second,
		MaxResponseHeaderBytes: 1 << 10,
		WriteBufferSize:        1 << 12,
		ReadBufferSize:         1 << 12,
		ForceAttemptHTTP2:      false,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return DefaultTransportWrapper(transport)
}

var (
	transportOnce                sync.Once
	transport                    http.RoundTripper
	insecureTransportOnce        sync.Once
	insecureTransport            http.RoundTripper
	wrappedTransportOnce         sync.Once
	wrappedTransport             http.RoundTripper
	wrappedInsecureTransportOnce sync.Once
	wrappedInsecureTransport     http.RoundTripper
)

func Transport(tws ...TransportWrapper) http.RoundTripper {
	transportOnce.Do(func() {
		transport = BuildTransport(tws...)
	})
	return transport
}

func InsecureTransport(tws ...TransportWrapper) http.RoundTripper {
	insecureTransportOnce.Do(func() {
		insecureTransport = BuildInsecureTransport(tws...)
	})
	return insecureTransport
}

func WrappedTransport() http.RoundTripper {
	wrappedTransportOnce.Do(func() {
		wrappedTransport = BuildWrappedTransport()
	})
	return wrappedTransport
}

func WrappedInsecureTransport() http.RoundTripper {
	wrappedInsecureTransportOnce.Do(func() {
		wrappedInsecureTransport = BuildWrappedInsecureTransport()
	})
	return wrappedInsecureTransport
}
