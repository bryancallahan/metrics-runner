// Metrics Runner (a simple data collection tool to gather analytics)
// Copyright (C) 2019  Bryan C. Callahan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package metricsrouter

import (
	"fmt"
	"log"
	"strings"
	"time"

	carbon "github.com/jforman/carbon-golang"

	"github.com/bryancallahan/metrics-runner/models"
)

type MetricsRouter struct {
	config         *models.Config
	carbonReceiver *carbon.Carbon
}

func connect(config *models.Config) (*carbon.Carbon, error) {
	return carbon.NewCarbon(config.MetricsRouter.CarbonHost, config.MetricsRouter.CarbonPort,
		!config.MetricsRouter.Enabled, config.MetricsRouter.Verbose)
}

// MetricsRouter initializes a structure and communicates for carbon stats reporting.
func NewMetricsRouter(config *models.Config) (*MetricsRouter, error) {

	carbonReceiver, err := connect(config)
	if err != nil {
		return &MetricsRouter{
			config:         config,
			carbonReceiver: nil,
		}, err
	}

	return &MetricsRouter{
		config:         config,
		carbonReceiver: carbonReceiver,
	}, nil
}

func (m *MetricsRouter) Write(path string, value float64) {

	// Note: github.com/jforman/carbon-golang is not ideal. It doesn't handle
	//  reconnections. So if grafana is restarted, all services that send to it
	//  would need to be restarted too. No client-side fibb backoff or client-side
	//  metric queuing. :'(
	//
	// To address this for now, we'll just try to reconnect if there's ANY kind
	//  of error and try again. If that doesn't succeed, just drop the metric. :'(

	now := time.Now().Unix()
	name := strings.Replace(strings.ToLower(m.config.Name), " ", "", -1)
	fullPath := fmt.Sprintf("%s-%s.%s", name, strings.ToLower(m.config.Env)[0:4], path)

	// If we don't have a receiver, don't send the metric...
	if m.carbonReceiver == nil {
		return
	}

	err := m.carbonReceiver.SendMetric(carbon.Metric{Name: fullPath, Value: value, Timestamp: now})
	if err != nil {

		// Attempt to reconnect...
		log.Println(fmt.Sprintf("error sending metric from metrics router: %s", err))
		log.Println(fmt.Sprintf("attempting to reconnect to %s:%d...", m.config.MetricsRouter.CarbonHost, m.config.MetricsRouter.CarbonPort))
		carbonReceiver, connErr := connect(m.config)
		if connErr != nil {
			log.Println(fmt.Sprintf("reconnect to %s:%d failed: %s (dropping metric for %s)", m.config.MetricsRouter.CarbonHost, m.config.MetricsRouter.CarbonPort, connErr, path))
			return
		}
		m.carbonReceiver = carbonReceiver // Save reconnection handle (for later use too)

		// Send metric again, if we can't then just drop it...
		log.Println(fmt.Sprintf("connection to %s:%d reestablished, resending metric for %s", m.config.MetricsRouter.CarbonHost, m.config.MetricsRouter.CarbonPort, path))
		errAttempt2 := m.carbonReceiver.SendMetric(carbon.Metric{Name: fullPath, Value: value, Timestamp: now})
		if errAttempt2 != nil {
			log.Println(fmt.Sprintf("error sending metric from metrics router: %s, dropping metric for %s", err, path))
			return
		}
	}
}
