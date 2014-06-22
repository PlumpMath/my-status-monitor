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

func main() {
	weatherChannel := weather.Monitor() // har har.
	batteryChannel := battery.Monitor("/sys/class/power_supply/BAT0/")
	ticker := time.NewTicker(time.Minute)
	timeFormat := "Mon Jan 2 3:04 PM"

	w, b := "", ""
	t := time.Now()
	tstring := t.Format(timeFormat)
	for {
		select {
		case w = <-weatherChannel:
		case b = <-batteryChannel:
		case t = <-ticker.C:
			tstring = t.Format(timeFormat)
		}

		setStatus(strings.Join([]string{w, b, tstring}, " | "))
	}
}
