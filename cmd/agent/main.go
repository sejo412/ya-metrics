package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
	"log"
	"sync"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var cfg Config
	pflag.StringVarP(&cfg.Address, "address", "a", DefaultServerAddress, "addressFlag to connect to")
	pflag.IntVarP(&cfg.ReportInterval, "reportInterval", "r", DefaultReportInterval, "report interval (in seconds)")
	pflag.IntVarP(&cfg.PollInterval, "pollInterval", "p", DefaultPollInterval, "poll interval (in seconds)")
	pflag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		return err
	}
	cfg.RealReportInterval = time.Duration(cfg.ReportInterval) * time.Second
	cfg.RealReportInterval = time.Duration(cfg.ReportInterval) * time.Second
	m := new(Metrics)
	r := new(Report)
	r.Gauge = make(map[string]float64)
	r.Counter = make(map[string]int64)
	var wg sync.WaitGroup
	wg.Add(1)
	go pollMetrics(m, &cfg)
	wg.Add(1)
	go reportMetrics(m, r, &cfg)
	wg.Wait()
	wg.Done()
	return nil
}
