package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-adoption-transaction/internal/transaction"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "run" {
		fatal("usage: gooo-adoption-transaction run --source PATH --contract PATH --out DIR")
	}
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	source := flags.String("source", "", "path to the .gooo source declaration")
	contract := flags.String("contract", "", "path to the fixed denominator contract")
	out := flags.String("out", "", "caller-owned empty output directory")
	if err := flags.Parse(os.Args[2:]); err != nil {
		os.Exit(2)
	}
	if *source == "" || *contract == "" || *out == "" {
		fatal("--source, --contract, and --out are required")
	}
	manifest, err := transaction.Run(*source, *contract, *out)
	if err != nil {
		fatal(err.Error())
	}
	data, err := json.Marshal(manifest.Summary)
	if err != nil {
		fatal(err.Error())
	}
	fmt.Println(string(data))
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
