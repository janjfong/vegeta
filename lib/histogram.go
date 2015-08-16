package vegeta

import (
	"fmt"
	"strings"
	"time"
)

// Buckets represent Histogram buckets.
type Buckets []time.Duration

// Histogram is bucketed latency Histogram.
type Histogram struct {
	Buckets Buckets
	Counts  []uint64
	Total   uint64
}

func NewHistogram(bs Buckets) Histogram {
	return Histogram{Buckets: bs, Counts: make([]uint64, len(bs))}
}

func (h *Histogram) Update(r Result) {
	var i int
	for ; i < len(h.Buckets)-1; i++ {
		if r.Latency >= h.Buckets[i] && r.Latency < h.Buckets[i+1] {
			break
		}
	}
	h.Total++
	h.Counts[i]++
}

// Nth returns the nth bucket represented as a string.
func (bs Buckets) Nth(i int) (left string, right string) {
	if i >= len(bs)-1 {
		return bs[i].String(), "+Inf"
	}
	return bs[i].String(), bs[i+1].String()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (bs *Buckets) UnmarshalText(value []byte) error {
	if len(value) < 2 || value[0] != '[' || value[len(value)-1] != ']' {
		return fmt.Errorf("bad buckets: %s", value)
	}
	for _, v := range strings.Split(string(value[1:len(value)-1]), ",") {
		d, err := time.ParseDuration(strings.TrimSpace(v))
		if err != nil {
			return err
		}
		*bs = append(*bs, d)
	}
	if len(*bs) == 0 {
		return fmt.Errorf("bad buckets: %s", value)
	}
	return nil
}
