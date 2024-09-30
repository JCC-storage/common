package json

import (
	"bytes"

	jsoniter "github.com/json-iterator/go"
)

type Serder struct {
	cfg Config
	api jsoniter.API
}

func (s *Serder) Encode(obj any) ([]byte, error) {
	buf := new(bytes.Buffer)

	enc := s.api.NewEncoder(buf)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *Serder) Decode(data []byte, obj any) error {
	dec := s.api.NewDecoder(bytes.NewReader(data))
	return dec.Decode(&obj)
}
