package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/sejo412/ya-metrics/cmd/agent/app"
	"github.com/spf13/pflag"
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
	pflag.BoolVarP(&cfg.UseOldApi, "oldApi", "o", DefaultUseOldApi, "use old api (deprecated)")
	pflag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		return err
	}
	cfg.RealReportInterval = time.Duration(cfg.ReportInterval) * time.Second
	cfg.RealReportInterval = time.Duration(cfg.ReportInterval) * time.Second
	m := new(app.Metrics)
	r := new(app.Report)
	r.Gauge = make(map[string]float64)
	r.Counter = make(map[string]int64)
	var wg sync.WaitGroup
	wg.Add(1)
	go app.PollMetrics(m, cfg.RealPollInterval)
	wg.Add(1)
	go app.ReportMetrics(m, r, fmt.Sprintf("%s://%s", ServerScheme, cfg.Address), cfg.RealReportInterval, ContextTimeout, cfg.UseOldApi)
	wg.Wait()
	wg.Done()
	return nil
}
