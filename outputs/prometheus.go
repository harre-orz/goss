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

	for resultGroup := range results {
		for _, testResult := range resultGroup {
			gossDurationSeconds.With(prometheus.Labels {
				"resource_type": testResult.ResourceType,
				"resource_id": testResult.ResourceId,
				"property": testResult.Property,
				"title": testResult.Title,
			}).Set(float64(testResult.Duration) / 1_000_000_000)
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
	registry.MustRegister(gossDurationSeconds, gossResult)
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
