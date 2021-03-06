package httpx

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// Codec Codec
type Codec interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
}

type jsonCodec struct{}

// JSONCodec JSONCodec
func JSONCodec() Codec {
	return &jsonCodec{}
}

func (c *jsonCodec) Encode(obj interface{}) ([]byte, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return data, nil
}

func (c *jsonCodec) Decode(data []byte, obj interface{}) error {
	err := json.Unmarshal(data, obj)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
