package agent

import (
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"math/rand/v2"
	"runtime"
	"sync/atomic"
)

type MetricProvider interface {
	GetMetrics() []common.MetricRequestDto
}

type RuntimeMetricProvider struct {
	pollCount int64
}

func (p *RuntimeMetricProvider) GetMetrics() []common.MetricRequestDto {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	atomic.AddInt64(&p.pollCount, 1)

	metrics := []common.MetricRequestDto{
		makeGaugeMetricRequest("Alloc", float64(memStats.Alloc)),
		makeGaugeMetricRequest("BuckHashSys", float64(memStats.BuckHashSys)),
		makeGaugeMetricRequest("GCCPUFraction", memStats.GCCPUFraction),
		makeGaugeMetricRequest("GCSys", float64(memStats.GCSys)),
		makeGaugeMetricRequest("HeapAlloc", float64(memStats.HeapAlloc)),
		makeGaugeMetricRequest("HeapIdle", float64(memStats.HeapIdle)),
		makeGaugeMetricRequest("HeapInuse", float64(memStats.HeapInuse)),
		makeGaugeMetricRequest("HeapObjects", float64(memStats.HeapObjects)),
		makeGaugeMetricRequest("HeapReleased", float64(memStats.HeapReleased)),
		makeGaugeMetricRequest("HeapSys", float64(memStats.HeapSys)),
		makeGaugeMetricRequest("LastGC", float64(memStats.LastGC)),
		makeGaugeMetricRequest("MCacheInuse", float64(memStats.MCacheInuse)),
		makeGaugeMetricRequest("MCacheSys", float64(memStats.MCacheSys)),
		makeGaugeMetricRequest("MSpanInuse", float64(memStats.MSpanInuse)),
		makeGaugeMetricRequest("MSpanSys", float64(memStats.MSpanSys)),
		makeGaugeMetricRequest("NextGC", float64(memStats.NextGC)),
		makeGaugeMetricRequest("Mallocs", float64(memStats.Mallocs)),
		makeGaugeMetricRequest("OtherSys", float64(memStats.OtherSys)),
		makeGaugeMetricRequest("StackInuse", float64(memStats.StackInuse)),
		makeGaugeMetricRequest("StackSys", float64(memStats.StackSys)),
		makeGaugeMetricRequest("Sys", float64(memStats.Sys)),
		makeGaugeMetricRequest("RandomValue", rand.Float64()),
		makeGaugeMetricRequest("Frees", float64(memStats.Frees)),
		makeGaugeMetricRequest("Lookups", float64(memStats.Lookups)),
		makeGaugeMetricRequest("NumForcedGC", float64(memStats.NumForcedGC)),
		makeGaugeMetricRequest("NumGC", float64(memStats.NumGC)),
		makeGaugeMetricRequest("PauseTotalNs", float64(memStats.PauseTotalNs)),
		makeGaugeMetricRequest("TotalAlloc", float64(memStats.TotalAlloc)),
		makeCounterMetricRequest("PollCount", p.pollCount),
	}

	return metrics
}

type VirtualMetricProvider struct {
	pollCount int64
}

func (p *VirtualMetricProvider) GetMetrics() []common.MetricRequestDto {
	v, _ := mem.VirtualMemory()
	cpuPercents, _ := cpu.Percent(0, true)

	metrics := []common.MetricRequestDto{
		makeGaugeMetricRequest("TotalMemory", float64(v.Total)),
		makeGaugeMetricRequest("FreeMemory", float64(v.Free)),
	}

	for i, percent := range cpuPercents {
		metrics = append(metrics, makeGaugeMetricRequest(fmt.Sprintf("CPUutilization%d", i+1), percent))
	}

	return metrics
}

func makeGaugeMetricRequest(id string, value float64) common.MetricRequestDto {
	return common.MetricRequestDto{
		ID:    id,
		MType: common.Gauge,
		Value: &value,
		Delta: nil,
	}
}

func makeCounterMetricRequest(id string, value int64) common.MetricRequestDto {
	return common.MetricRequestDto{
		ID:    id,
		MType: common.Counter,
		Value: nil,
		Delta: &value,
	}
}
