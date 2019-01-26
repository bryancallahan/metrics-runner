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
	"encoding/json"
	"fmt"
	"time"
)

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {

	var rawValue interface{}
	err := json.Unmarshal(b, &rawValue)
	if err != nil {
		return err
	}

	// Attempt to unmarshal if float64 or string...
	switch value := rawValue.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil

	case string:
		duration, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		d.Duration = duration
		return nil

	default:
		return fmt.Errorf("unable to parse duration")
	}
}
