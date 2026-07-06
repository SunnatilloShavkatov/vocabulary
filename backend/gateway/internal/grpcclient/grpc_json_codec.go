package grpcclient

import "encoding/json"

type grpcJSONCodec struct{}

func (grpcJSONCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (grpcJSONCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (grpcJSONCodec) Name() string {
	return "json"
}
