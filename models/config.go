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

package models

import (
	"github.com/jmoiron/jsonq"

	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type ConfigMetricsRouter struct {
	Enabled    bool   `json:"enabled"`
	Verbose    bool   `json:"verbose"`
	CarbonHost string `json:"carbonHost"`
	CarbonPort int    `json:"carbonPort"`
}

type ConfigMetric struct {
	Type        string   `json:"type"` // e.g. "build-number", "http"
	Name        string   `json:"name"`
	Method      string   `json:"method"`
	URL         string   `json:"url"`
	Periodicity Duration `json:"periodicity"` // Need to use our Duration so we can unmarshal
}

type Config struct {
	Env           string              `json:"env"`
	TLSEnable     bool                `json:"tlsEnable"`
	TLSCRT        string              `json:"tlsCRT"`
	TLSKey        string              `json:"tlsKey"`
	Port          int                 `json:"port"`
	LogFile       string              `json:"logFile"`
	Name          string              `json:"name"`
	MetricsRouter ConfigMetricsRouter `json:"metricsRouter"`
	Metrics       []ConfigMetric      `json:"metrics"`
}

func NewConfig() (*Config, error) {

	// Grab the environmental variable....
	env := os.Getenv("GO_ENV")
	if len(env) < 1 {
		env = "development"
	}

	// Load development config (development config is special, prod is loaded
	// first and then dev is loaded on top as overrides) or production / test
	// configs (which are taken literally)...
	if env == "development" {
		return NewDevelopmentConfig()
	} else {
		return NewConfigUsing(fmt.Sprintf("config-%s.json", env))
	}
}

func NewConfigUsing(configFile string) (*Config, error) {

	c := &Config{}
	err := decodeJson(configFile, c)
	if err != nil {
		return &Config{}, err
	}

	return c, nil
}

func NewDevelopmentConfig() (*Config, error) {

	c := &Config{}

	// Load the production config first...
	err := decodeJson("config-production.json", c)
	if err != nil {
		return &Config{}, err
	}

	// Load the development config second (so anything it has overrides prod)...
	err = decodeJson("config-development.json", c)
	if err != nil {
		return &Config{}, err
	}

	return c, nil
}

func (c *Config) LogSummary() {
	log.Println("configuration summary")
	log.Println(" environment: ........", c.Env)
	log.Println(" tls enabled: ........", fmt.Sprintf("%t", c.TLSEnable))
	log.Println(" access log file: ....", c.LogFile)
	log.Println(" name: ...............", c.Name)
}

// ReadProperty lets properties be read from the config using a query
// path (e.g. "metrics.0.name"). If any errors occur, we return
// an empty string.
func (c *Config) ReadProperty(queryPath string) string {

	data := map[string]interface{}{}

	// Convert the current config struct into a format jsonq can query...
	b, _ := json.Marshal(c)
	decoder := json.NewDecoder(strings.NewReader(string(b)))
	decoder.Decode(&data)

	// Break the query path into a string array that jsonq can work with...
	jq := jsonq.NewQuery(data)
	value, _ := jq.String(strings.Split(queryPath, ".")...)

	return value
}

func decodeJson(filename string, s *Config) error {

	// Open the file...
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the json using given interface...
	decoder := json.NewDecoder(file)
	err = decoder.Decode(s)
	if err != nil {
		return err
	}

	// Make sure all metric names are unique...
	countMap := map[string]int{}
	for _, metric := range s.Metrics {
		countMap[metric.Name]++
		if countMap[metric.Name] > 1 {
			return fmt.Errorf("found duplicate metric by the name of %s (please make sure all "+
				"configured metric names are unique)", metric.Name)
		}

		if len(metric.Name) < 1 {
			return fmt.Errorf("found empty metric name (please make sure all metrics have a name property)")
		}
	}

	return nil
}
