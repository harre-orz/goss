package outputs

import (
	"io"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

type Prometheus struct{}

func (r Prometheus) Output(w io.Writer, results <-chan []resource.TestResult,
	startTime time.Time, outConfig util.OutputConfig) (exitCode int) {

	var gossDurationSeconds, gossResult *prometheus.GaugeVec
	gossDurationSeconds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "goss_duration_seconds",
		Help: "Lets you know duration of goss execution",
	},[]string{"resource_type", "resource_id", "property", "title"})
	gossResult = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "goss_result",
		Help: "Lets you know if goss assertions were true 0, or false 1, or skip 2",
	},[]string{"resource_type", "resource_id", "property", "title"})

	var gossTestsCount, gossTestsFailedCount, gossTestsSkippedCount prometheus.Gauge
	gossTestsCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "goss_tests_count",
		Help: "Test count of goss assertions",
	})
	gossTestsFailedCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "goss_tests_failed_count",
		Help: "Test failed count of goss assertions",
	})
	gossTestsSkippedCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "goss_tests_skipped_count",
		Help: "Test skipped count of goss assertions",
	})

	for resultGroup := range results {
		for _, testResult := range resultGroup {
			switch testResult.Result {
			case resource.FAIL:
				gossTestsFailedCount.Inc()
			case resource.SKIP:
				gossTestsSkippedCount.Inc()
			}
			gossTestsCount.Inc()
			gossDurationSeconds.With(prometheus.Labels {
				"resource_type": testResult.ResourceType,
				"resource_id": testResult.ResourceId,
				"property": testResult.Property,
				"title": testResult.Title,
			}).Set(testResult.Duration.Seconds())
			gossResult.With(prometheus.Labels {
				"resource_type": testResult.ResourceType,
				"resource_id": testResult.ResourceId,
				"property": testResult.Property,
				"title": testResult.Title,
			}).Set(float64(testResult.Result))
		}
	}

	var registry *prometheus.Registry
	registry = prometheus.NewRegistry()
	registry.MustRegister(gossDurationSeconds, gossResult,
		gossTestsCount, gossTestsFailedCount, gossTestsSkippedCount)
	mfs, err := registry.Gather()
	if err != nil {
		return 1
	}
	for _, mf := range mfs {
		expfmt.MetricFamilyToText(w, mf)
	}
	return 0
}

func init() {
	RegisterOutputer("prometheus", &Prometheus{}, []string{})
}
