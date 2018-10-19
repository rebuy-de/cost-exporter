package prom

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func Run(port string) {
	r := prometheus.NewRegistry()
	r.MustRegister(C.CoreCount)
	r.MustRegister(C.Cost)
	r.MustRegister(C.ReservationCoverage)
	r.MustRegister(C.SpotRequest)

	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))

	go func() {
		logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
	}()
}
