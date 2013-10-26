package main

import (
	"os/exec"
	"time"
)

const (
	timeFormat = "Mon Jan 2 3:04 PM"
)

func emit(t time.Time) {
	cmd := exec.Command("xsetroot", "-name", t.Format(timeFormat))
	cmd.Run()
}

func main() {
	for {
		t := time.Now()
		emit(t)
		time.Sleep(time.Second)
	}
}
