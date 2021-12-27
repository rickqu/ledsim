package metrics

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartMetrics() {
	go runHeartbeat()
}

func runHeartbeat() {
	heartbeat := true
	heartbeatMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "heartbeat_square_wave",
		Help: "A square wave to show heartbeat.",
	})
	prometheus.MustRegister(heartbeatMetric)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":80", nil))
	}()
	for {
		if heartbeat {
			heartbeatMetric.Set(1.0)
		} else {
			heartbeatMetric.Set(0.0)
		}
		heartbeat = !heartbeat
		time.Sleep(31e9)
	}
}
