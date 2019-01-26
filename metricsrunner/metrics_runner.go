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

package metricsrunner

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/bryancallahan/metrics-runner/metricsrouter"
	"github.com/bryancallahan/metrics-runner/models"
	"github.com/bryancallahan/metrics-runner/utilities"
)

type MetricsRunner struct {
	sync.Mutex

	wg            sync.WaitGroup
	config        *models.Config
	version       *models.Version
	metricsRouter *metricsrouter.MetricsRouter
	metric        *models.ConfigMetric
}

func NewMetricsRunner(version *models.Version, config *models.Config,
	metricsRouter *metricsrouter.MetricsRouter, metric models.ConfigMetric) *MetricsRunner {
	return &MetricsRunner{
		config:        config,
		version:       version,
		metricsRouter: metricsRouter,
		metric:        &metric,
	}
}

func (m *MetricsRunner) run() error {

	m.wg.Add(1)
	defer m.wg.Done()

	m.Lock()
	defer m.Unlock()

	switch m.metric.Type {
	case "build-number":
		// Send the build number (note: if short hash ends with "-dev", we're running uncommitted
		//  code, show this using half-steps)...
		buildNumber := float64(m.version.BuildNumber)
		if strings.HasSuffix(m.version.ShortHash, "-dev") {
			buildNumber += 0.5
		}
		m.metricsRouter.Write(fmt.Sprintf("%s.%s", m.metric.Type, m.metric.Name), buildNumber)
		return nil

	case "http":

		// For the moment we only support GETs...
		if m.metric.Method != "GET" {
			return fmt.Errorf("method %s is currently not supported", m.metric.Method)
		}

		elapsed, statusCode, _, err := utilities.GetURL(false, m.metric.URL)
		if err != nil {
			return err
		}
		log.Println(fmt.Sprintf("%s %s - Elapsed: %s, Status Code: %d", m.metric.Method, m.metric.URL, elapsed, statusCode))
		m.metricsRouter.Write(fmt.Sprintf("%s.%s.elapsed", m.metric.Type, m.metric.Name), float64(elapsed/time.Microsecond)/1000.0)
		m.metricsRouter.Write(fmt.Sprintf("%s.%s.status-code", m.metric.Type, m.metric.Name), float64(statusCode))
		return nil

	default:
		return fmt.Errorf("could not run metrics runner for type %s as it is an unsupported type", m.metric.Type)
	}
}

func (m *MetricsRunner) Start() {

	startDelay := 2 + (int)(10*rand.Float64())
	log.Print(fmt.Sprintf("waiting %d seconds before starting metrics runner for %s", startDelay, m.metric.Name))
	time.Sleep(time.Duration(startDelay) * time.Second)
	log.Print(fmt.Sprintf("metrics runner started for %s with a periodicity of %s", m.metric.Name, m.metric.Periodicity))
	m.metricsRouter.Write("started", 1)

	for {

		err := m.run()
		if err != nil {
			log.Println("metricsrunner:", err)
		}

		time.Sleep(m.metric.Periodicity.Duration)
	}
}

func (m *MetricsRunner) Stop() error {

	const stopTimeout = 30 // Seconds

	c := make(chan struct{}, 1)
	go func() {
		m.wg.Wait()
		c <- struct{}{}
	}()

	select {
	case <-c:
		return nil
	case <-time.After(stopTimeout * time.Second):
		return fmt.Errorf("metricsrunner: wait group did not finish within %d seconds", stopTimeout)
	}
}

// Metric returns a copy of the metric currently executing by this metrics runner.
func (m *MetricsRunner) Metric() *models.ConfigMetric {
	return &models.ConfigMetric{
		Type:        m.metric.Type,
		Name:        m.metric.Name,
		Method:      m.metric.Method,
		URL:         m.metric.URL,
		Periodicity: m.metric.Periodicity,
	}
}
