package httpx

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
)

func DrainBody(src io.ReadCloser) ([]byte, io.ReadCloser, error) {
	defer src.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(src); err != nil {
		return nil, nil, err
	}
	if buf.Len() > 0 {
		return buf.Bytes(), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
	}
	return nil, ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
