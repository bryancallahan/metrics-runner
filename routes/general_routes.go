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

package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/bryancallahan/metrics-runner/models"
	"github.com/bryancallahan/metrics-runner/utilities"
)

const PASSWORD_RESET_TOKEN_LENGTH = 64

func InitializeGeneralRoutes(version *models.Version, config *models.Config, r *mux.Router) {
	apiRouter := r.PathPrefix("/api/").Subrouter()
	apiRouter.HandleFunc("/heartbeat", newGetHeartbeat(version, config)).Methods("GET")
	apiRouter.HandleFunc("/config", newGetConfig(version, config)).Methods("GET")
}

func newGetHeartbeat(version *models.Version, config *models.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		utilities.ServeJSON(w, r, http.StatusOK, nil)
	}
}

func newGetConfig(version *models.Version, config *models.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		utilities.ServeJSON(w, r, http.StatusOK, map[string]interface{}{
			"version": *version,
		})
	}
}
