package collector

import (
	"github.com/skaes/logjam-tools/go/util"
	"sync"
	"testing"
)

func TestExtractingMetricNames(t *testing.T) {
	metric, kind := extractLogjamMetricFromName("logjam:action:db_time_summary_seconds")
	if metric != "db_time" {
		t.Errorf("could not extract metric name %s", "db_time")
	}
	if kind != "summary" {
		t.Errorf("could not extract metric kind %s", "summary")
	}
	metric, kind = extractLogjamMetricFromName("logjam:action:db_time_distribution_seconds")
	if metric != "db_time" {
		t.Errorf("could not extract metric name %s", "db_time")
	}
	if kind != "distribution" {
		t.Errorf("could not extract metric kind %s", "summary")
	}
	metric, kind = extractLogjamMetricFromName("logjam:action:db_time_murks_seconds")
	if metric != "" {
		t.Errorf("should return exmpty string when extracting metric name")
	}
	metric, kind = extractLogjamMetricFromName("logjam:action:db_calls_total")
	if metric != "db_calls" {
		t.Errorf("could not extract metric name %s", "db_calls")
	}
	if kind != "total" {
		t.Errorf("could not extract metric kind %s", "total")
	}
}

func TestDeletingLabels(t *testing.T) {
	s := util.Stream{
		App:                 "a",
		Env:                 "b",
		IgnoredRequestURI:   "/_",
		BackendOnlyRequests: "",
		APIRequests:         []string{},
	}
	options := Options{
		Datacenters: "a,b",
		CleanAfter:  60,
		Resources: &util.Resources{
			TimeResources: []string{"db_time"},
			CallResources: []string{"db_calls"},
		},
	}
	c := New(s.AppEnv(), &s, options)
	metrics1 := &metric{
		kind: logMetric,
		props: map[string]string{
			"app":     "a",
			"env":     "b",
			"metric":  "http",
			"code":    "200",
			"method":  "GET",
			"cluster": "c",
			"dc":      "d",
			"action":  "murks",
		},
		value:          5.7,
		timeMetrics:    map[string]float64{"db_time": 1.45},
		counterMetrics: map[string]float64{"db_calls": 1},
	}
	metrics2 := &metric{
		kind: logMetric,
		props: map[string]string{
			"app":     "a",
			"env":     "b",
			"metric":  "http",
			"code":    "200",
			"method":  "GET",
			"cluster": "d",
			"dc":      "e",
			"action":  "marks",
		},
		value:          7.7,
		timeMetrics:    map[string]float64{"db_time": 1.45},
		counterMetrics: map[string]float64{"db_calls": 1},
	}
	metrics3 := &metric{
		kind: logMetric,
		props: map[string]string{
			"app":     "a",
			"env":     "b",
			"metric":  "job",
			"code":    "200",
			"cluster": "d",
			"dc":      "e",
			"action":  "marks",
		},
		value:          3.1,
		timeMetrics:    map[string]float64{"db_time": 1.45},
		counterMetrics: map[string]float64{"db_calls": 1},
	}
	metrics4 := &metric{
		kind: logMetric,
		props: map[string]string{
			"app":     "a",
			"env":     "b",
			"metric":  "job",
			"code":    "200",
			"cluster": "d",
			"dc":      "e",
			"action":  "marks"},
		value:          4.4,
		timeMetrics:    map[string]float64{"db_time": 1.45},
		counterMetrics: map[string]float64{"db_calls": 1},
	}
	c.recordLogMetrics(metrics1)
	c.recordLogMetrics(metrics2)
	c.recordLogMetrics(metrics3)
	c.recordLogMetrics(metrics4)
	if !c.removeAction("marks") {
		t.Errorf("could not remove action: %s", "marks")
	}
	if c.removeAction("schnippi") {
		t.Errorf("could remove non existing action : %s", "schnippi")
	}
}

var m sync.Mutex
var res int

func lockUnlock(n int) int {
	m.Lock()
	defer m.Unlock()
	return n + 1
}

func BenchmarkLocking(b *testing.B) {
	r := 0
	for i := 0; i < b.N; i++ {
		r = lockUnlock(r)
	}
	res = r
}

var l sync.RWMutex

func rwLockUnlock(n int) int {
	l.RLock()
	defer l.RUnlock()
	return n + 1
}

func BenchmarkRWLocking(b *testing.B) {
	r := 0
	for i := 0; i < b.N; i++ {
		r = rwLockUnlock(r)
	}
	res = r
}

func BenchmarkRecordingLogMetrics(b *testing.B) {
	s := util.Stream{
		App:                 "a",
		Env:                 "b",
		IgnoredRequestURI:   "/_",
		BackendOnlyRequests: "",
		APIRequests:         []string{},
	}
	options := Options{
		Datacenters: "a,b",
		CleanAfter:  60,
	}
	c := New(s.AppEnv(), &s, options)
	for i := 0; i < b.N; i++ {
		metrics := &metric{
			kind: logMetric,
			props: map[string]string{
				"app":     "a",
				"env":     "b",
				"metric":  "http",
				"code":    "200",
				"method":  "GET",
				"cluster": "c",
				"dc":      "d",
				"action":  "murks",
			},
			value: 5.7,
		}
		c.recordLogMetrics(metrics)
	}
}

// $ go test -run=XXX -bench=.
// goos: darwin
// goarch: amd64
// pkg: github.com/skaes/logjam-tools/go/prometheusexporter/collector
// BenchmarkLocking-12					34417060			29.9 ns/op
// BenchmarkRWLocking-12				38610462			30.1 ns/op
// BenchmarkRecordingLogMetrics-12		  363196		  3044 ns/op
// PASS
// ok	github.com/skaes/logjam-tools/go/prometheusexporter/collector	3.416s

// This means metrics processing slows down by roughly 1% after adding locking.
