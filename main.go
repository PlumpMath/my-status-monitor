package main

import (
	"fmt"
	"os/exec"
	"time"

	"strings"

	"github.com/zenhack/my-status-monitor/battery"
)

func emit(str string) {
	cmd := exec.Command("xsetroot", "-name", str)
	cmd.Run()
}

func getWeather() (string, error) {
	cmd := exec.Command("show-weather.sh")
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func main() {
	timeFormat := "Mon Jan 2 3:04 PM"
	myBattery := &battery.Battery{Path: "/sys/class/power_supply/BAT0/"}

	go myBattery.Monitor()
	go myBattery.NotifyDaemon()

	for {
		t := time.Now()
		batteryState := myBattery.State
		weather, weatherErr := getWeather()
		state := t.Format(timeFormat)
		if batteryState != nil {
			state = fmt.Sprintf("%v | %s", batteryState, state)
		}
		if weatherErr == nil {
			state = fmt.Sprintf("%s | %s", weather, state)
		}
		emit(state)
		time.Sleep(time.Second)
	}
}
