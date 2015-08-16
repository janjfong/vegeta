package main

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/tsenart/vegeta/ioutil"
	vegeta "github.com/tsenart/vegeta/lib"
)

func reportCmd() command {
	fs := flag.NewFlagSet("vegeta report", flag.ExitOnError)
	reporter := fs.String("reporter", "text", "Reporter [text, json, plot, hist[buckets]]")
	window := fs.Duration("window", time.Second, "Batch window")
	inputs := fs.String("inputs", "stdin", "Input files (comma separated)")
	output := fs.String("output", "stdout", "Output file")
	return command{fs, func(args []string) error {
		fs.Parse(args)
		return report(*reporter, *inputs, *output, *window)
	}}
}

// report validates the report arguments, sets up the required resources
// and writes the report every window duration
func report(reporter, inputs, output string, window time.Duration) error {
	if len(reporter) < 4 {
		return fmt.Errorf("bad reporter: %s", reporter)
	}
	files := strings.Split(inputs, ",")
	srcs := make([]io.Reader, len(files))
	for i, f := range files {
		in, err := file(f, false)
		if err != nil {
			return err
		}
		defer in.Close()
		srcs[i] = in
	}

	out, err := file(output, true)
	if err != nil {
		return err
	}
	defer out.Close()

	dec := vegeta.NewDecoder(ioutil.RoundRobinReader(srcs...))

	var report vegeta.Reporter
	switch reporter[:4] {
	case "text":
		report = vegeta.NewTextReporter(dec, window)
	case "json":
		report = vegeta.NewJSONReporter(dec, window)
	case "plot":
		report = vegeta.NewPlotReporter(dec)
	case "hist":
		if len(reporter) < 6 {
			return fmt.Errorf("bad buckets: '%s'", reporter[4:])
		}
		var bs vegeta.Buckets
		if err := bs.UnmarshalText([]byte(reporter[4:])); err != nil {
			return err
		}
		report = vegeta.NewHistogramReporter(dec, bs, window)
	}
	return report(out)
}
