package carbon

import (
	"fmt"
	"net"
	"time"
)

type Carbon struct {
	Host     string
	Port     int
	Timeout  time.Duration
	conn     net.Conn
	Noop     bool
	Protocol string
	Verbose  bool
}

// time in Seconds of how long each Carbon transaction is allowed to take
const defaultTimeout = 5

func (carbon *Carbon) IsNoop() bool {
	if carbon.Noop {
		return true
	}
	return false
}

// Is a Carbon Host defined? In other words, do we intend to send to carbon?
func (carbon *Carbon) IsDefined() bool {
	if carbon.Host == "" {
		return false
	}
	return true
}

func (carbon *Carbon) Connect() error {
	if carbon.IsDefined() {
		if carbon.conn != nil {
			carbon.conn.Close()
		}
		if !carbon.IsNoop() {
			address := fmt.Sprintf("%s:%d", carbon.Host, carbon.Port)
			fmt.Printf("Carbon Reciever: %v.\n", address)
			conn, err := net.DialTimeout(carbon.Protocol, address, carbon.Timeout)
			if err != nil {
				return err
			}
			carbon.conn = conn
		}
	}
	return nil
}

// Given a carbon Metric, send it to Carbon.
func (carbon *Carbon) SendMetric(metric Metric) error {
	_, err := fmt.Fprintf(carbon.conn, metric.String()+"\n")
	if carbon.Verbose {
		fmt.Printf("in SendMetrics, metric.String: %s\n", metric.String())
	}
	if err != nil {
		return err
	}
	return nil
}

// Given a list of metrics, send them each invidually.
func (carbon *Carbon) SendMetrics(metrics []Metric) error {
	if carbon.IsDefined() {
		if carbon.IsNoop() {
			return nil
		}
		start := time.Now()
		for _, metric := range metrics {
			err := carbon.SendMetric(metric)
			if err != nil {
				return err
			}
		}
		elapsed := time.Since(start)
		if carbon.Verbose {
			fmt.Printf("Elapsed time to post to Carbon %v.\n", elapsed)
		}
	}
	return nil
}

// Create new Carbon object as well as connect to Carbon instance.
func NewCarbon(host string, port int, noop bool, verbose bool) (*Carbon, error) {
	carbon := &Carbon{Host: host, Port: port, Timeout: time.Duration(1 * time.Minute), Protocol: "tcp", Noop: noop,
		Verbose: verbose}
	err := carbon.Connect()
	if err != nil {
		return nil, err
	}
	return carbon, nil
}
