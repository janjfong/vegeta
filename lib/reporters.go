package vegeta

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

// Reporter is a function type that represents Result reporters.
type Reporter func(io.Writer) error

// NewHistogramReporter returns a Reporter that computes latency histograms with the
// given buckets.
func NewHistogramReporter(dec Decoder, bs Buckets, window time.Duration) Reporter {
	return func(w io.Writer) (err error) {
		var r Result
		h := NewHistogram(bs)
		tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', tabwriter.StripEscape)
		ticker := time.NewTicker(window)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Fprintf(tw, "Bucket\t\t#\t%%\tHistogram\n")
				for i, count := range h.Counts {
					ratio := float64(count) / float64(h.Total)
					lo, hi := h.Buckets.Nth(i)
					pad := strings.Repeat("#", int(ratio*75))
					fmt.Fprintf(tw, "[%s,\t%s]\t%d\t%.2f%%\t%s\n", lo, hi, count, ratio*100, pad)
					if err := tw.Flush(); err != nil {
						return err
					}
				}
			default:
				if err = dec(&r); err != nil {
					return err
				}
				h.Update(r)
			}
		}
		return tw.Flush()
	}
}

// NewTextReporter returns a Reporter which computes Metrics from the given
// Results and writes them to the given io.Writer as aligned, formatted text.
func NewTextReporter(dec Decoder, window time.Duration) Reporter {
	return func(w io.Writer) error {
		var r Result
		var m Metrics
		tw := tabwriter.NewWriter(w, 0, 8, 2, '\t', tabwriter.StripEscape)
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fmt.Fprintf(tw, "Requests\t[total, rate]\t%d, %.2f\n", m.Requests, m.Rate)
				fmt.Fprintf(tw, "Duration\t[total, attack, wait]\t%s, %s, %s\n", m.Duration+m.Wait, m.Duration, m.Wait)
				fmt.Fprintf(tw, "Latencies\t[mean, 50, 95, 99, max]\t%s, %s, %s, %s, %s\n",
					m.Latencies.Mean, m.Latencies.P50, m.Latencies.P95, m.Latencies.P99, m.Latencies.Max)
				fmt.Fprintf(tw, "Bytes In\t[total, mean]\t%d, %.2f\n", m.BytesIn.Total, m.BytesIn.Mean)
				fmt.Fprintf(tw, "Bytes Out\t[total, mean]\t%d, %.2f\n", m.BytesOut.Total, m.BytesOut.Mean)
				fmt.Fprintf(tw, "Success\t[ratio]\t%.2f%%\n", m.Success*100)
				fmt.Fprintf(tw, "Status Codes\t[code:count]\t")
				for code, count := range m.StatusCodes {
					fmt.Fprintf(tw, "%s:%d  ", code, count)
				}
				fmt.Fprintln(tw, "\nError Set:")
				for _, err := range m.Errors {
					fmt.Fprintln(tw, err)
				}

				if err := tw.Flush(); err != nil {
					return err
				}
			default:
				if err := dec(&r); err != nil {
					return err
				}
				m.Update(r)
			}
		}
		return tw.Flush()
	}
}

// NewJSONReporter returns a Reporter which computes Metrics from the given
// Results and writes them to the given io.Writer as JSON.
func NewJSONReporter(dec Decoder, window time.Duration) Reporter {
	return func(w io.Writer) (err error) {
		var r Result
		var m Metrics
		enc := json.NewEncoder(w)
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err = enc.Encode(m); err != nil {
					return err
				}
			default:
				if err = dec(&r); err != nil {
					return err
				}
				m.Update(r)
			}
		}
		return nil
	}
}

// NewPlotReporter returns a Reporter which writes a self contained HTML
// page with an interactive plot of the latencies of the requests.
// Built with http://dygraphs.com/
// BUG(tsenart): Verify if results need to be written sorted or not.
func NewPlotReporter(dec Decoder) Reporter {
	return func(w io.Writer) error {
		_, err := fmt.Fprintf(w, plotsTemplateHead, asset(dygraphs), asset(html2canvas))
		if err != nil {
			return err
		}

		var r Result
		buf := make([]byte, 0, 128)
		for {
			if err = dec(&r); err != nil {
				return err
			}
			buf = append(buf, '[')
			buf = append(buf, r.Timestamp.String()...)
			buf = append(buf, ',')

			latency := strconv.FormatFloat(r.Latency.Seconds()*1000, 'f', -1, 32)
			if r.Error == "" {
				buf = append(buf, "NaN,"...)
				buf = append(buf, latency...)
				buf = append(buf, ']', ',')
			} else {
				buf = append(buf, latency...)
				buf = append(buf, ",NaN],"...)
			}

			if _, err := w.Write(buf[:len(buf)-1]); err != nil {
				return err
			}
			buf = buf[:0]
		}

		_, err = fmt.Fprint(w, plotsTemplateTail)
		return err
	}
}

const (
	plotsTemplateHead = `<!doctype html>
<html>
<head>
  <title>Vegeta Plots</title>
</head>
<body>
  <div id="latencies" style="font-family: Courier; width: 100%%; height: 600px"></div>
  <button id="download">Download as PNG</button>
  <script>%s</script>
  <script>%s</script>
  <script>
  new Dygraph(
    document.getElementById("latencies"),
    [`
	plotsTemplateTail = `],
    {
      title: 'Vegeta Plot',
      labels: ['Seconds', 'ERR', 'OK'],
      ylabel: 'Latency (ms)',
      xlabel: 'Seconds elapsed',
      showRoller: true,
      colors: ['#FA7878', '#8AE234'],
      legend: 'always',
      logscale: true,
      strokeWidth: 1.3
    }
  );
  document.getElementById("download").addEventListener("click", function(e) {
    html2canvas(document.body, {background: "#fff"}).then(function(canvas) {
      var url = canvas.toDataURL('image/png').replace(/^data:image\/[^;]/, 'data:application/octet-stream');
      var a = document.createElement("a");
      a.setAttribute("download", "vegeta-plot.png");
      a.setAttribute("href", url);
      a.click();
    });
  });
  </script>
</body>
</html>`
)
