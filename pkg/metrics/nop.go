package metrics

type nopMetricsCollector struct{}

func NewNopMetricsCollector() nopMetricsCollector {
	return nopMetricsCollector{}
}

func (m nopMetricsCollector) CacheHit(method string, statusCode int) {
}

func (m nopMetricsCollector) CacheMiss(method string, statusCode int) {
}
