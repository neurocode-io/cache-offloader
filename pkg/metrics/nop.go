package metrics

type NopMetricsCollector struct{}

func NewNopMetricsCollector() NopMetricsCollector {
	return NopMetricsCollector{}
}

func (m NopMetricsCollector) CacheHit(method string, statusCode int) {
}

func (m NopMetricsCollector) CacheMiss(method string, statusCode int) {
}
