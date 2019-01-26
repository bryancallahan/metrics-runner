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

package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"

	"github.com/bryancallahan/metrics-runner/models"
)

// NewLoggingHandler creates a new Logging Handler for access log.
func NewLoggingHandler(config *models.Config) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {

		file, err := os.OpenFile(config.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Println(fmt.Errorf("could not initialize logging handler: %s", err))
		}

		return handlers.LoggingHandler(file, h)
	}
}
