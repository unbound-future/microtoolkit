package metrics

import "github.com/prometheus/client_golang/prometheus"

func RegisterMetrics(registry prometheus.Registerer) {
	registry.MustRegister(errLogTotalCounter)
	registry.MustRegister(channelDispatchCounter)
}

var (
	errLogTotalCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "err_log_total",
			Help: "Total number of error log",
		}, []string{"err_log_type"})
	channelDispatchCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "channel_dispatch_total",
			Help: "Total number of times a channel is dispatched",
		},
		[]string{"channel_id", "model_name", "priority", "weight", "retry_times"},
	)
)

func IncrementErrLogTotalCounter(errLogType string, add float64) {
	errLogTotalCounter.WithLabelValues(errLogType).Add(add)
}

func IncrementChannelDispatchCounter(channelID, modelName, priority, weight, retryTimes string, add float64) {
	channelDispatchCounter.WithLabelValues(channelID, modelName, priority, weight, retryTimes).Add(add)
}
