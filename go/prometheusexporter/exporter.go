package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	log "github.com/skaes/logjam-tools/go/logging"
	"github.com/skaes/logjam-tools/go/prometheusexporter/collector"
	"github.com/skaes/logjam-tools/go/prometheusexporter/collectormanager"
	"github.com/skaes/logjam-tools/go/prometheusexporter/messageparser"
	"github.com/skaes/logjam-tools/go/prometheusexporter/stats"
	"github.com/skaes/logjam-tools/go/prometheusexporter/webserver"
	"github.com/skaes/logjam-tools/go/util"
)

var opts struct {
	Verbose       bool   `short:"v" long:"verbose" description:"Verbose logging."`
	StatsInterval int    `short:"r" long:"stats-interval" default:"1" description:"Number of seconds between reporting statistics."`
	LogjamURL     string `short:"l" long:"logjam-url" env:"LOGJAM_URL" default:"http://localhost:3000" description:"Logjam instance to use for retrieving stream definitions."`
	Devices       string `short:"d" long:"devices" env:"LOGJAM_DEVICES" default:"127.0.0.1:9606,127.0.0.1:9706" description:"Comma separated device specs (host:port pairs)."`
	Env           string `short:"e" long:"env" env:"LOGJAM_ENV" description:"Logjam environments to process."`
	Datacenters   string `short:"D" long:"datacenters" env:"LOGJAM_DATACENTERS" description:"List of known datacenters, comma separated. Will be used to determine label value if not available on incoming data."`
	DefaultDC     string `short:"u" long:"default-dc" env:"LOGJAM_DATACENTER" default:"unknown" description:"Assume this datacenter name if none could be derived from incoming data."`
	Parsers       uint   `short:"P" long:"parsers" default:"4" description:"Number of message parsers to run in parallel."`
	CleanAfter    uint   `short:"c" long:"clean-after" default:"5" description:"Minutes to wait before cleaning old time series."`
	Port          string `short:"p" long:"port" default:"8081" description:"Port to expose metrics on."`
	AbortAfter    uint   `short:"A" long:"abort" env:"LOGJAM_ABORT_AFTER" default:"60" description:"Abort after missing heartbeats for this many seconds."`
	RcvHWM        int    `short:"R" long:"rcv-hwm" env:"LOGJAM_RCV_HWM" default:"1000000" description:"Zmq high water mark for receive socket."`
}

func parseArgs() {
	args, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		e := err.(*flags.Error)
		if e.Type != flags.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
	if len(args) > 1 {
		log.Error("%s: passing arguments is not supported. please use options instead.", args[0])
		os.Exit(1)
	}
}

func main() {
	log.Info("%s starting", os.Args[0])
	parseArgs()

	util.InstallSignalHandler()

	collectorOptions := collector.Options{
		Verbose:     opts.Verbose,
		Debug:       false,
		Datacenters: opts.Datacenters,
		DefaultDC:   opts.DefaultDC,
		CleanAfter:  opts.CleanAfter,
	}
	collectormanager.Initialize(opts.LogjamURL, opts.Env, collectorOptions)

	go stats.ReporterAndWatchdog(opts.AbortAfter, opts.Verbose, opts.StatsInterval)

	parserOptions := messageparser.Options{
		Verbose: opts.Verbose,
		Debug:   false,
		Parsers: opts.Parsers,
		Devices: opts.Devices,
		RcvHWM:  opts.RcvHWM,
	}
	go messageparser.New(parserOptions).Run()

	webserver.HandleHTTPRequests(opts.Port)

	log.Info("%s shutdown", os.Args[0])
}
