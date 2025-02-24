package json

import (
	// prov "github.com/bytedance/sonic"
	prov "github.com/goccy/go-json"
	"io"
)

func Unmarshal(data []byte, v any) error {
	return prov.Unmarshal(data, v)
}

func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return prov.MarshalIndent(v, prefix, indent)
}

func Marshal(v any) ([]byte, error) {
	return prov.Marshal(v)
}

// Encoder encodes JSON into io.Writer
type Encoder interface {
	Encode(val any) error
}

type Decoder interface {
	Decode(val any) error
}

func NewEncoder(w io.Writer) Encoder {
	//return prov.ConfigDefault.NewEncoder(w)
	return prov.NewEncoder(w)
}

func NewDecoder(r io.Reader) Decoder {
	//return prov.ConfigDefault.NewDecoder(r)
	return prov.NewDecoder(r)
}
