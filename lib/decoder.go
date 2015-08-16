package vegeta

import (
	"encoding/gob"
	"io"
)

func init() {
	gob.Register(&Result{})
}

// A Decoder decodes Results from an io.Reader
type Decoder func(*Result) error

// NewDecoder returns a new Decoder for the given io.Reader
func NewDecoder(r io.Reader) Decoder {
	dec := gob.NewDecoder(r)
	return func(r *Result) error {
		return dec.Decode(r)
	}
}
