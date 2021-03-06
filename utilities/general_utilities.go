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

package utilities

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func WaitForOSSignal(terminate func() error) {

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {

		sig := <-signalChannel
		log.Println("received", sig, "signal from OS")

		var err error

		err = terminate()

		log.Println("terminating")
		if err == nil {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}()
}

// ServeJSON writes status and response as JSON.
func ServeJSON(w http.ResponseWriter, r *http.Request, status int, response interface{}) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if response == nil {
		response = struct{}{}
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Could not encode json response", err)
	}
}
