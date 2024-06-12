package agent

import (
	"github.com/desepticon55/metrics-collector/internal/common"
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync/atomic"
)

type MetricProvider interface {
	GetMetrics() []common.Metric
}

type RuntimeMetricProvider struct {
	pollCount int64
}

func (p *RuntimeMetricProvider) GetMetrics() []common.Metric {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	atomic.AddInt64(&p.pollCount, 1)

	metrics := []common.Metric{
		{Name: "Alloc", Value: strconv.FormatUint(memStats.Alloc, 10), Type: common.Gauge},
		{Name: "BuckHashSys", Value: strconv.FormatUint(memStats.BuckHashSys, 10), Type: common.Gauge},
		{Name: "Frees", Value: strconv.FormatUint(memStats.Frees, 10), Type: common.Counter},
		{Name: "GCCPUFraction", Value: strconv.FormatFloat(memStats.GCCPUFraction, 'f', -1, 64), Type: common.Gauge},
		{Name: "GCSys", Value: strconv.FormatUint(memStats.GCSys, 10), Type: common.Gauge},
		{Name: "HeapAlloc", Value: strconv.FormatUint(memStats.HeapAlloc, 10), Type: common.Gauge},
		{Name: "HeapIdle", Value: strconv.FormatUint(memStats.HeapIdle, 10), Type: common.Gauge},
		{Name: "HeapInuse", Value: strconv.FormatUint(memStats.HeapInuse, 10), Type: common.Gauge},
		{Name: "HeapObjects", Value: strconv.FormatUint(memStats.HeapObjects, 10), Type: common.Gauge},
		{Name: "HeapReleased", Value: strconv.FormatUint(memStats.HeapReleased, 10), Type: common.Gauge},
		{Name: "HeapSys", Value: strconv.FormatUint(memStats.HeapSys, 10), Type: common.Gauge},
		{Name: "LastGC", Value: strconv.FormatUint(memStats.LastGC, 10), Type: common.Gauge},
		{Name: "Lookups", Value: strconv.FormatUint(memStats.Lookups, 10), Type: common.Counter},
		{Name: "MCacheInuse", Value: strconv.FormatUint(memStats.MCacheInuse, 10), Type: common.Gauge},
		{Name: "MCacheSys", Value: strconv.FormatUint(memStats.MCacheSys, 10), Type: common.Gauge},
		{Name: "MSpanInuse", Value: strconv.FormatUint(memStats.MSpanInuse, 10), Type: common.Gauge},
		{Name: "MSpanSys", Value: strconv.FormatUint(memStats.MSpanSys, 10), Type: common.Gauge},
		{Name: "Mallocs", Value: strconv.FormatUint(memStats.Mallocs, 10), Type: common.Counter},
		{Name: "NextGC", Value: strconv.FormatUint(memStats.NextGC, 10), Type: common.Gauge},
		{Name: "NumForcedGC", Value: strconv.FormatUint(uint64(memStats.NumForcedGC), 10), Type: common.Counter},
		{Name: "NumGC", Value: strconv.FormatUint(uint64(memStats.NumGC), 10), Type: common.Counter},
		{Name: "OtherSys", Value: strconv.FormatUint(memStats.OtherSys, 10), Type: common.Gauge},
		{Name: "PauseTotalNs", Value: strconv.FormatUint(memStats.PauseTotalNs, 10), Type: common.Counter},
		{Name: "StackInuse", Value: strconv.FormatUint(memStats.StackInuse, 10), Type: common.Gauge},
		{Name: "StackSys", Value: strconv.FormatUint(memStats.StackSys, 10), Type: common.Gauge},
		{Name: "Sys", Value: strconv.FormatUint(memStats.Sys, 10), Type: common.Gauge},
		{Name: "TotalAlloc", Value: strconv.FormatUint(memStats.TotalAlloc, 10), Type: common.Counter},
		{Name: "PollCount", Value: strconv.FormatInt(p.pollCount, 10), Type: common.Counter},
		{Name: "RandomValue", Value: strconv.FormatFloat(rand.Float64(), 'f', -1, 64), Type: common.Gauge},
	}

	return metrics
}
