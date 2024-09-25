package agent

import "fmt"

// Example to use RuntimeMetricProvider
func ExampleRuntimeMetricProvider_GetMetrics() {
	provider := &RuntimeMetricProvider{}

	metrics := provider.GetMetrics()

	for _, metric := range metrics {
		if metric.Value != nil {
			fmt.Printf("Metric: %s, Value: %.2f\n", metric.ID, *metric.Value)
		} else if metric.Delta != nil {
			fmt.Printf("Metric: %s, Delta: %d\n", metric.ID, *metric.Delta)
		}
	}
}

// Example to use VirtualMetricProvider
func ExampleVirtualMetricProvider_GetMetrics() {
	provider := &VirtualMetricProvider{}

	metrics := provider.GetMetrics()

	for _, metric := range metrics {
		if metric.Value != nil {
			fmt.Printf("Metric: %s, Value: %.2f\n", metric.ID, *metric.Value)
		} else if metric.Delta != nil {
			fmt.Printf("Metric: %s, Delta: %d\n", metric.ID, *metric.Delta)
		}
	}
}
