package vegeta

import (
	"encoding/json"
	"fmt"
	"io"
)

// Dumper is a function type that represents Result dumpers.
type Dumper func(io.Writer) error

// NewCSVDumper returns a Dumper that dumps each incoming Result as a CSV
// record with six columns. The columns are: unix timestamp in ns since epoch,
// http status code, request latency in ns, bytes out, bytes in, and lastly the error.
func NewCSVDumper(dec Decoder) Dumper {
	return func(w io.Writer) (err error) {
		var r Result
		for {
			if err = dec(&r); err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "%d,%d,%d,%d,%d,\"%s\"\n",
				r.Timestamp.UnixNano(),
				r.Code,
				r.Latency.Nanoseconds(),
				r.BytesOut,
				r.BytesIn,
				r.Error,
			)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// NewJSONDumper returns a Dumper with dumps each incoming Result as a JSON object.
func NewJSONDumper(dec Decoder) Dumper {
	return func(w io.Writer) (err error) {
		var r Result
		enc := json.NewEncoder(w).Encode
		for {
			if err = dec(&r); err != nil {
				return err
			} else if err = enc(r); err != nil {
				return err
			}
		}
		return nil
	}
}
