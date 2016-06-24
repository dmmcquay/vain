package metrics

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Errors tracks http status codes for problematic requests.
	Errors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Number of upstream errors",
		},
		[]string{"status"},
	)

	// Func tracks time spent in a function.
	Func = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "function_microseconds",
			Help: "function timing.",
		},
		[]string{"route"},
	)

	// DB tracks timing of interactions with the file system.
	DB = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "db_microseconds",
			Help: "db timing.",
		},
		[]string{"what"},
	)
)

func init() {
	prometheus.MustRegister(Errors)
	prometheus.MustRegister(Func)
	prometheus.MustRegister(DB)
}

// Time is a function that makes it simple to add one-line timings to function
// calls.
func Time() func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start)
		pc := make([]uintptr, 10)
		runtime.Callers(2, pc)
		f := runtime.FuncForPC(pc[0])

		Func.WithLabelValues(f.Name()).Observe(float64(elapsed / time.Microsecond))
	}
}

// DBTime makes it simple to add one-line timings to db interactions.
func DBTime(name string) func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start)
		DB.WithLabelValues(name).Observe(float64(elapsed / time.Microsecond))
	}
}
