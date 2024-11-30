package main

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func testConfig() *Config {
	return &Config{
		RealReportInterval: 2 * time.Second,
		RealPollInterval:   1 * time.Second,
	}
}

func Test_parseMetric(t *testing.T) {
	type args struct {
		root       *string
		metricName *string
		data       reflect.Value
		report     *Report
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parseMetric(tt.args.root, tt.args.metricName, tt.args.data, tt.args.report)
		})
	}
}

func Test_pollMetrics(t *testing.T) {
	type args struct {
		m *Metrics
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pollMetrics(tt.args.m, testConfig())
		})
	}
}

func Test_postMetric(t *testing.T) {
	type args struct {
		ctx    context.Context
		metric string
		ch     chan string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postMetric(tt.args.ctx, tt.args.metric, testConfig(), tt.args.ch)
		})
	}
}

func Test_reportMetrics(t *testing.T) {
	type args struct {
		m      *Metrics
		report *Report
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportMetrics(tt.args.m, tt.args.report, testConfig())
		})
	}
}
