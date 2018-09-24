package prom

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Costs struct {
	Cost                prometheus.GaugeVec
	CoreCount           prometheus.GaugeVec
	ReservationCoverage prometheus.GaugeVec
}

var C = Costs{
	Cost: *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "rebuy",
			Subsystem: "cost_exporter",
			Name:      "costs",
			Help:      "Costs by account and by service.",
		},
		[]string{"account", "service"},
	),
	CoreCount: *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "rebuy",
			Subsystem: "cost_exporter",
			Name:      "cores",
			Help:      "Count of all virtual CPUs in all regions of a specific account.",
		},
		[]string{"account", "region"},
	),
	ReservationCoverage: *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "rebuy",
			Subsystem: "cost_exporter",
			Name:      "reservationcoverage",
			Help:      "Coverage of running EC2 instances by reservations in percent.",
		},
		[]string{"account"},
	),
}

func (c *Costs) SetTotalCoreCount(account string, region string, count float64) {
	c.CoreCount.With(prometheus.Labels{
		"account": account,
		"region":  region,
	}).Set(count)
}

func (c *Costs) SetCosts(account string, service string, cost float64) {
	c.Cost.With(prometheus.Labels{
		"account": account,
		"service": service,
	}).Set(cost)
}

func (c *Costs) SetReservationCoverage(account string, coverage float64) {
	c.ReservationCoverage.With(prometheus.Labels{
		"account": account,
	}).Set(coverage)
}
