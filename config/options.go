package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var ErrNotAllowed = errors.New("input is not allowed")
var ErrNoArgument = errors.New("argument is not set")
var version = "v0.1.0"
var Banner = `
                _
__ ___ __  _ __| |
\ \ / '  \| '_ \ |
/_\_\_|_|_| .__/_|
          |_|
            ` + version + `
`

var usage = Banner + "\n" + `
Usage: example [options]
Options:
	-w,  --worker-count, *[Default: 100]
	-k,  --kafka         *[Usage: brokers<>inTopic<>outToic. Example: localhost:9094,localhost:9095<>in<>out]	
	-n,  --no-banner
	-v,  --verbose
	-h,  --help

* means "must be set".`

func PrintErrorAndDie(err error) {
	fmt.Println(usage)
	color.Red(err.Error())
	os.Exit(1)
}

func printHelpAndExit() {
	fmt.Println(usage)
	os.Exit(0)
}

type Options struct {
	WorkerCount int
	Kafka       string
	Brokers     []string
	InTopic     string
	OutTopic    string
	NoBanner    bool
	Verbose     bool
	ShowHelp    bool
}

// Configure sets options for the server.
func Configure(fs *flag.FlagSet, args []string) (*Options, error) {
	opts := &Options{}
	fs.Usage = func() {}

	fs.IntVar(&opts.WorkerCount, "w", 100, "")
	fs.IntVar(&opts.WorkerCount, "worker-count", 100, "")
	fs.StringVar(&opts.Kafka, "k", "", "")
	fs.StringVar(&opts.Kafka, "kafka", "", "")
	fs.BoolVar(&opts.NoBanner, "n", false, "")
	fs.BoolVar(&opts.NoBanner, "no-banner", false, "")
	fs.BoolVar(&opts.Verbose, "v", false, "")
	fs.BoolVar(&opts.Verbose, "verbose", false, "")
	fs.BoolVar(&opts.ShowHelp, "h", false, "")
	fs.BoolVar(&opts.ShowHelp, "help", false, "")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if opts.ShowHelp {
		printHelpAndExit()
	}
	if opts.WorkerCount < 1 {
		fmt.Println(usage)
		return nil, fmt.Errorf("%w, workerCount must be posive", ErrNotAllowed)
	}

	kafkas := strings.Split(opts.Kafka, "<>")
	if len(kafkas) != 3 {
		return nil, fmt.Errorf("%w, kafka usage is wrong, example: %s", ErrNotAllowed, "localhost:9094<>in<>out")
	}

	opts.Brokers = strings.Split(kafkas[0], ",")
	opts.InTopic = kafkas[1]
	opts.OutTopic = kafkas[2]

	if len(opts.Brokers) <= 0 {
		return nil, fmt.Errorf("%w, brokers must be set", ErrNoArgument)
	}
	if opts.InTopic == "" || opts.OutTopic == "" {
		return nil, fmt.Errorf("%w, in-out topics must be set", ErrNoArgument)
	}

	return opts, nil
}
