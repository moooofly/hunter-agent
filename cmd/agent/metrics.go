package main

import (
	"expvar"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	metrics "github.com/docker/go-metrics"
	"github.com/sirupsen/logrus"
)

var (
	drop   = expvar.NewInt("overflow-drop")
	msgIn  = expvar.NewInt("message-in")
	msgOut = expvar.NewInt("message-out")
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

var defaultPeriod = 5 * time.Second

// TODO(moooofly): support period setting by user
func logExpvars() {
	logrus.Debugf("Metrics logging every %s", defaultPeriod)

	ticker := time.NewTicker(defaultPeriod)
	prevVals := map[string]int64{}
	for {
		<-ticker.C
		vals := map[string]int64{}
		snapshotExpvars(vals)
		metrics := buildMetricsOutput(prevVals, vals)
		prevVals = vals
		if len(metrics) > 0 {
			logrus.Debugf("Non-zero metrics in the last %s:%s", defaultPeriod, metrics)
		} else {
			logrus.Debugf("No non-zero metrics in the last %s", defaultPeriod)
		}
	}
}

func logTotalExpvars() {
	vals := map[string]int64{}
	prevVals := map[string]int64{}
	snapshotExpvars(vals)
	metrics := buildMetricsOutput(prevVals, vals)
	logrus.Debugf("Total non-zero values: %s", metrics)
	logrus.Debugf("Uptime: %s", time.Now().Sub(startTime))
}

// snapshotMap recursively walks expvar Maps and records their integer expvars
// in a separate flat map.
func snapshotMap(varsMap map[string]int64, path string, mp *expvar.Map) {
	mp.Do(func(kv expvar.KeyValue) {
		switch kv.Value.(type) {
		case *expvar.Int:
			varsMap[path+"."+kv.Key], _ = strconv.ParseInt(kv.Value.String(), 10, 64)
		case *expvar.Map:
			snapshotMap(varsMap, path+"."+kv.Key, kv.Value.(*expvar.Map))
		}
	})
}

// snapshotExpvars iterates through all the defined expvars, and for the vars
// that are integers it snapshots the name and value in a separate (flat) map.
func snapshotExpvars(varsMap map[string]int64) {
	expvar.Do(func(kv expvar.KeyValue) {
		switch kv.Value.(type) {
		case *expvar.Int:
			varsMap[kv.Key], _ = strconv.ParseInt(kv.Value.String(), 10, 64)
		case *expvar.Map:
			snapshotMap(varsMap, kv.Key, kv.Value.(*expvar.Map))
		}
	})
}

// buildMetricsOutput makes the delta between vals and prevVals and builds
// a printable string with the non-zero deltas.
func buildMetricsOutput(prevVals map[string]int64, vals map[string]int64) string {
	metrics := ""
	for k, v := range vals {
		delta := v - prevVals[k]
		if delta != 0 {
			metrics = fmt.Sprintf("%s %s=%d", metrics, k, delta)
		}
	}
	return metrics
}

func startMetricsServer(addr string, w chan error) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.Handler())

	go func() {
		if err := http.Serve(l, mux); err != nil {
			logrus.Errorf("serve metrics api: %s", err)
			w <- nil
		}
	}()
	return nil
}
