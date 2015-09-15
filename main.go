package main

import (
	"os/exec"
	"strings"
	"time"

	"github.com/zenhack/my-status-monitor/battery"
	"github.com/zenhack/my-status-monitor/weather"
)

func setStatus(str string) {
	cmd := exec.Command("xsetroot", "-name", str)
	cmd.Run()
}

// Like strings.Join(strs, " | "), but filters out empty strings in strs first.
func join(strs ...string) string {
	resultSlice := make([]string, 0, len(strs))
	for i := range(strs) {
		if strs[i] != "" {
			resultSlice = append(resultSlice, strs[i])
		}
	}
	return strings.Join(resultSlice, " | ")
}

func main() {
	weatherChannel := weather.Monitor() // har har.
	batteryChannel := battery.Monitor("/sys/class/power_supply/BAT0/")
	ticker := time.NewTicker(time.Minute)
	timeFormat := "Mon Jan 2 3:04 PM"

	w, b := "", ""
	t := time.Now()
	tstring := t.Format(timeFormat)
	for {
		setStatus(join(w, b, tstring))

		select {
		case w = <-weatherChannel:
		case b = <-batteryChannel:
		case t = <-ticker.C:
			tstring = t.Format(timeFormat)
		}
	}
}
