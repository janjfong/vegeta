package main

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/tsenart/vegeta/ioutil"
	vegeta "github.com/tsenart/vegeta/lib"
)

func dumpCmd() command {
	fs := flag.NewFlagSet("vegeta dump", flag.ExitOnError)
	dumper := fs.String("dumper", "json", "Dumper [json, csv]")
	inputs := fs.String("inputs", "stdin", "Input files (comma separated)")
	output := fs.String("output", "stdout", "Output file")
	return command{fs, func(args []string) error {
		fs.Parse(args)
		return dump(*dumper, *inputs, *output)
	}}
}

func dump(dumper, inputs, output string) error {
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

	var dump vegeta.Dumper
	switch dumper {
	case "csv":
		dump = vegeta.NewCSVDumper(dec)
	case "json":
		dump = vegeta.NewJSONDumper(dec)
	default:
		return fmt.Errorf("unsupported dumper: %s", dumper)
	}

	return dump(out)
}
