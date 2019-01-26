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
	"log"
	"net/http"

	throttled "gopkg.in/throttled/throttled.v2"
	"gopkg.in/throttled/throttled.v2/store/memstore"
)

// ThrottleHandler controls the number of requests that should
// be throttled to the server.
func ThrottleHandler(h http.Handler) http.Handler {

	throttleStore, err := memstore.New(65536)
	if err != nil {
		log.Fatal(err)
	}

	return throttled.RateLimit(throttled.PerMin(30),
		&throttled.VaryBy{RemoteAddr: true},
		throttleStore).Throttle(h)
}
