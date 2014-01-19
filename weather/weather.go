package weather

import (
	"os/exec"
	"strings"
	"time"
)

func fetch() error {
	cmd := exec.Command("get-weather.sh")
	return cmd.Run()
}

func show() string {
	cmd := exec.Command("show-weather.sh")
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

func Monitor() <-chan string {
	ch := make(chan string, 1)
	fetch()
	go func() {
		for {
			ch <- show()
			time.Sleep(20 * time.Minute)
			fetch()
		}
	}()
	return ch
}
