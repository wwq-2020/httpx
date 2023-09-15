package httpx

import (
	"encoding/json"
	"fmt"
	"io"
)

type Codec interface {
	Decode(io.Reader, interface{}) error
	Encode(interface{}) ([]byte, error)
}

type codec struct {
	encoder func(interface{}) ([]byte, error)
	decoder func(io.Reader, interface{}) error
}

func NewCodec(encoder func(interface{}) ([]byte, error),
	decoder func(io.Reader, interface{}) error) Codec {
	return &codec{
		encoder: encoder,
		decoder: decoder,
	}
}

func (c *codec) Encode(obj interface{}) ([]byte, error) {
	return c.encoder(obj)
}

func (c *codec) Decode(r io.Reader, obj interface{}) error {
	return c.decoder(r, obj)
}

type JsonCodec struct {
}

var defaultCodec = &JsonCodec{}

func (c *JsonCodec) Decode(r io.Reader, obj interface{}) error {
	if err := json.NewDecoder(r).Decode(obj); err != nil {
		return err
	}
	return nil
}

func (c *JsonCodec) Encode(obj interface{}) ([]byte, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type StatusJsonCodec struct{}

type statusResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (c *StatusJsonCodec) Decode(r io.Reader, obj interface{}) error {
	resp := &statusResp{
		Data: obj,
	}
	if err := json.NewDecoder(r).Decode(resp); err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("unexpected code:%d,msg:%s", resp.Code, resp.Msg)
	}
	return nil
}

func (c *StatusJsonCodec) Encode(obj interface{}) ([]byte, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return data, nil
}
