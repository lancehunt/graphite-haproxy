/*
   conntrack-logger
   Copyright (C) 2015 Denis V Chapligin <akashihi@gmail.com>
   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"fmt"
	"strconv"
	"time"
)

func wait(configuration Configuration) {
	time.Sleep(time.Duration(configuration.Period) * time.Second)
}

func main() {
	InitLog()
	log.Info("Starting graphite-haproxy...")

	var configuration = config()

	for {
		var page, err = getPage(configuration.StatusUrl)
		if err != nil {
			wait(configuration)
			continue
		}
		status, err := parse(page)
		if err != nil {
			wait(configuration)
			continue
		}

		fmt.Println("Parsed!")
		// modify the status slice to extrapolate calculated fields
		computeSyntheticFields(status)
		fmt.Println("Computed!")

		sendMetrics(status, configuration)
		fmt.Println("Sent!")
		wait(configuration)
	}
}

func computeSyntheticFields2(status []Status) {
	for item := range status {
		fmt.Println(item)
	}
}

var last map[string]Status

func computeSyntheticFields(status []Status) {
	// create delta for EReq, ECon, EResp, total_connections

	if last == nil {
		last = make(map[string]Status)
	}

	for i, _ := range status {
		if lastStatus, ok := last[status[i].Name]; ok {
			status[i].EReqRate = diffInt(status[i].EReq, lastStatus.EReq)
			status[i].EConRate = diffInt(status[i].ECon, lastStatus.ECon)
			status[i].ERespRate = diffInt(status[i].EResp, lastStatus.EResp)

		} else {
			status[i].EReqRate = status[i].EReq
			status[i].EConRate = diffInt(status[i].ECon, lastStatus.ECon)
			status[i].ERespRate = diffInt(status[i].EResp, lastStatus.EResp)
		}
	}

	last = mapItemsByName(status)
}

func diffInt(current, last string) string {
	currentI, err := strconv.Atoi(current)
	if err != nil {
		return ""
	}
	lastI, err := strconv.Atoi(last)
	if err != nil {
		lastI = 0
	}
	return strconv.Itoa(currentI - lastI)
}

func mapItemsByName(status []Status) map[string]Status {
	var items = map[string]Status{}

	for _, item := range status {
		items[item.Name] = item
	}

	return items
}

func mapItemsByType(status []Status, itemType string) map[string]Status {
	var items = map[string]Status{}

	for _, item := range status {
		switch item.Type {
		case itemType:
			items[item.Name] = item
			continue
		}
	}

	return items
}
