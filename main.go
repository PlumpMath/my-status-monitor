package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/zenhack/my-status-monitor/battery"
)

const (
)

func emit(str string) {
	cmd := exec.Command("xsetroot", "-name", str)
	cmd.Run()
}


func main() {
	timeFormat := "Mon Jan 2 3:04 PM"
	myBattery := &battery.Battery{Path: "/sys/class/power_supply/BAT0/"}

	go myBattery.Monitor()

	for {
		t := time.Now()
		state := myBattery.State
		if state != nil {
			emit(fmt.Sprintf("%v | %s", state, t.Format(timeFormat)))
		} else {
			emit(t.Format(timeFormat))
		}
		time.Sleep(time.Second)
	}
}
