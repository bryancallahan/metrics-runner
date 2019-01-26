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
	"os"
)

type Version struct {
	BuildNumber int    `json:"buildNumber"`
	Hash        string `json:"hash"`
	ShortHash   string `json:"shortHash"`
}

func NewVersion() (*Version, error) {

	// Open the version file...
	file, err := os.Open("version.json")
	if err != nil {
		return &Version{}, err
	}
	defer file.Close()

	// Decode the version information...
	decoder := json.NewDecoder(file)
	version := &Version{}
	err = decoder.Decode(version)
	if err != nil {
		return &Version{}, err
	}

	return version, nil
}

func (v *Version) BuildHash() string {
	return fmt.Sprintf("%d-%s", v.BuildNumber, v.ShortHash)
}
