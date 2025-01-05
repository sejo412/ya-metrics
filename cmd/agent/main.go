package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/sejo412/ya-metrics/internal/app/agent"
	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/spf13/pflag"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var cfg config.AgentConfig
	pflag.StringVarP(&cfg.Address, "address", "a", config.DefaultServerAddress, "addressFlag to connect to")
	pflag.IntVarP(&cfg.ReportInterval, "reportInterval", "r", config.DefaultReportInterval,
		"report interval (in seconds)")
	pflag.IntVarP(&cfg.PollInterval, "pollInterval", "p", config.DefaultPollInterval, "poll interval (in seconds)")
	pflag.BoolVarP(&cfg.UseOldAPI, "oldApi", "o", config.DefaultUseOldAPI, "use old api (deprecated)")
	pflag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		return err
	}
	cfg.RealReportInterval = time.Duration(cfg.ReportInterval) * time.Second
	cfg.RealReportInterval = time.Duration(cfg.ReportInterval) * time.Second
	m := new(agent.Metrics)
	r := new(agent.Report)
	r.Gauge = make(map[string]float64)
	r.Counter = make(map[string]int64)
	var wg sync.WaitGroup
	wg.Add(1)
	go agent.PollMetrics(m, cfg.RealPollInterval)
	wg.Add(1)
	go agent.ReportMetrics(m, r, fmt.Sprintf("%s://%s", config.ServerScheme, cfg.Address), cfg.RealReportInterval,
		config.ContextTimeout, cfg.UseOldAPI)
	wg.Wait()
	wg.Done()
	return nil
}
