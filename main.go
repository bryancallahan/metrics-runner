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

package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"

	"github.com/bryancallahan/metrics-runner/metricsrouter"
	"github.com/bryancallahan/metrics-runner/metricsrunner"
	"github.com/bryancallahan/metrics-runner/middleware"
	"github.com/bryancallahan/metrics-runner/models"
	"github.com/bryancallahan/metrics-runner/routes"
	"github.com/bryancallahan/metrics-runner/utilities"
)

var (
	version *models.Version
	config  *models.Config

	metricsRouter  *metricsrouter.MetricsRouter
	metricsRunners []*metricsrunner.MetricsRunner
)

func terminate() error {

	var err error
	var hasError bool

	if len(metricsRunners) > 0 {
		log.Println(fmt.Sprintf("stopping %d metrics runners", len(metricsRunners)))
		for _, metricsRunner := range metricsRunners {
			log.Println(fmt.Sprintf("stopping metrics runner for %s", metricsRunner.Metric().Name))
			err = metricsRunner.Stop()
			if err != nil {
				log.Println(fmt.Sprintf("error stopping metrics runner for %s: %s", metricsRunner.Metric().Name, err))
				hasError = true
			}
		}
	}

	if hasError {
		return fmt.Errorf("error attempting to cleanly terminate")
	}

	return nil
}

func main() {

	// Configure executable flags...
	doesHideTimestampFlag := flag.Bool("hidetimestamp", false, "hides the timestamp in logs (useful when being invoked from systemd)")
	readPropertyFlag := flag.String("readproperty", "", "reads a property from json configuration (observes GO_ENV)")
	flag.Parse()

	// If true, we run in "read property mode" (we should roll up all of this flag stuff
	// so that it isn't sitting in main)...
	hasReadProperty := (len(*readPropertyFlag) > 0)

	// Configure logger flags...
	if *doesHideTimestampFlag {
		log.SetFlags(0)
	} else {
		log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	}

	// Load up the version...
	v, _ := models.NewVersion()
	version = v

	// Load up the config...
	c, err := models.NewConfig()
	if err != nil {
		if !hasReadProperty {
			log.Println(fmt.Sprintf("error could not load configuration: %s", err))
		}
		os.Exit(1)
	}
	config = c

	// If we have a read property, read the value from the config and exit...
	if hasReadProperty {
		fmt.Println(config.ReadProperty(*readPropertyFlag))
		os.Exit(0)
	}

	// Start your engines...
	log.Println(fmt.Sprintf("Metrics Runner, %s (a simple data collection tool to gather analytics)", v.BuildHash()))
	log.Println("Copyright (C) 2019  Bryan C. Callahan")
	log.Println("")
	log.Println("This program is free software: you can redistribute it and/or modify")
	log.Println("it under the terms of the GNU General Public License as published by")
	log.Println("the Free Software Foundation, either version 3 of the License, or")
	log.Println("(at your option) any later version.")
	log.Println("")
	log.Println("This program is distributed in the hope that it will be useful,")
	log.Println("but WITHOUT ANY WARRANTY; without even the implied warranty of")
	log.Println("MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the")
	log.Println("GNU General Public License for more details.")
	log.Println("")
	log.Println("You should have received a copy of the GNU General Public License")
	log.Println("along with this program.  If not, see <https://www.gnu.org/licenses/>.")
	log.Println("")
	config.LogSummary()
	log.Println("")

	// Seed the random number generator...
	rand.Seed(time.Now().UTC().UnixNano())

	// Start up metrics router...
	metricsRouter, err = metricsrouter.NewMetricsRouter(config)
	if err != nil {
		log.Println(fmt.Sprintf("error initializing metrics router: %s", err))
	}

	// Start all the metrics runners...
	for _, metric := range config.Metrics {

		if !metric.Enabled {
			log.Println(fmt.Sprintf("skipping metrics runner for %s", metric.Name))
			continue
		}

		metricsRunner := metricsrunner.NewMetricsRunner(version, config, metricsRouter, metric)
		go metricsRunner.Start()
		metricsRunners = append(metricsRunners, metricsRunner)
	}

	// Start waiting for term signals (handles things like proper cleanup on interrupt / term)...
	utilities.WaitForOSSignal(terminate)

	// Create the router for the primary API...
	r := mux.NewRouter()
	routes.InitializeGeneralRoutes(version, config, r)

	// Assemble all middleware and create master handler...
	http.Handle("/", context.ClearHandler(alice.New(middleware.ThrottleHandler,
		middleware.NewLoggingHandler(config),
		handlers.CompressHandler).Then(r)))

	// Start up the api...
	log.Println(fmt.Sprintf("started api (listening on *:%d)", config.Port))
	if config.TLSEnable {

		httpServer := models.NewSecureServer(config)
		err = httpServer.ListenAndServeTLS(config.TLSCRT, config.TLSKey)
		if err != nil {
			log.Fatal("error listening to TLS connections", err)
		}

	} else {

		httpServer := models.NewServer(config)
		err = httpServer.ListenAndServe()
		if err != nil {
			log.Fatal("error listening to connections", err)
		}
	}
}
