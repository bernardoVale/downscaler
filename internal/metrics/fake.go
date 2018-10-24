package metrics

import (
	"io/ioutil"
	"strings"
	"time"
)

// fakeMetrics struct fake ingress metrics using a newline file where each line you have
// the full app name and how much time the app is idle separated by a comma.
// E.g: foo/bar,10s
type fakeMetrics struct {
	metricsFile string
}

// NewFakeMetricsClient creates a metrics Client based on a newline file
func NewFakeMetricsClient(metricsFile string) Client {
	return fakeMetrics{metricsFile}
}

func (f fakeMetrics) MetricExists(metricName string) (bool, error) {
	return true, nil
}

func (f fakeMetrics) ListActiveIngresses(maxIdle string) (map[string]bool, error) {
	activeIngresses := make(map[string]bool)

	maxIdleDuration, err := time.ParseDuration(maxIdle)
	if err != nil {
		return activeIngresses, err
	}

	data, err := ioutil.ReadFile(f.metricsFile)
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		lineData := strings.Split(line, ",")
		if len(lineData) != 2 {
			continue
		}
		app := lineData[0]
		idleTime := lineData[1]

		duration, err := time.ParseDuration(idleTime)
		if err != nil {
			return activeIngresses, err
		}
		if maxIdleDuration >= duration {
			activeIngresses[app] = true
		}
	}
	return activeIngresses, nil
}
