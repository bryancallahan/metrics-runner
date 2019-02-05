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
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// NewServer creates a new HTTP Server (nonTLS).
func NewServer(config *Config) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// NewSecureServer creates a new Secure HTTP Server (TLS).
func NewSecureServer(config *Config) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS10,
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			},
		},
	}
}

func QueryHTTPMetric(metric *ConfigMetric) (time.Duration, int, []byte, bool, error) {

	// We can only query metrics of http type...
	if metric.Type != "http" {
		return 0, 0, nil, false, fmt.Errorf("cannot query metric type %s via http", metric.Type)
	}

	// Validate supported methods...
	validMethods := map[string]bool{
		"GET":  true,
		"POST": true,
	}
	if !validMethods[metric.Method] {
		return 0, 0, nil, false, fmt.Errorf("method %s is currently not supported", metric.Method)
	}

	// Configure transport (allow us to optionally ignore bad certs)...
	cookieJar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false}, // TODO - Pull InsecureSkipVerify from config
	}
	client := &http.Client{
		Jar:       cookieJar,
		Transport: transport,
		Timeout:   metric.Timeout.Duration,
	}

	// Bind data map to form data...
	// TODO - Load Content-Type from config and act accordingly here (i.e. we should be able
	//   to send application/json payloads -- or any payloads for that matter).
	form := url.Values{}
	for key, value := range metric.Data {
		form.Add(key, value)
	}

	// Create request...
	// TODO - Load Content-Type from config and act accordingly here (i.e. we should be able
	//   to send application/json payloads -- or any payloads for that matter).
	req, err := http.NewRequest(metric.Method, metric.URL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return 0, 0, nil, false, err
	}

	// Add optional headers...
	for key, value := range metric.Headers {
		req.Header.Add(key, value)
	}

	start := time.Now()

	// Make request...
	res, err := client.Do(req)
	if err != nil {
		return 0, 0, nil, false, err
	}
	defer res.Body.Close()

	// Read the whole body...
	var body []byte
	if res.Body != nil {
		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return 0, 0, nil, false, err
		}
	}

	elapsed := time.Since(start)

	// Check if we're valid...
	isValid := true
	if res.StatusCode != 200 {
		isValid = false
	}
	if len(metric.StringToCheck) > 0 && !strings.Contains(string(body), metric.StringToCheck) {
		isValid = false
	}

	return elapsed, res.StatusCode, body, isValid, nil
}
