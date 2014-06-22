package weather

import (
	"os/exec"
	"strings"
	"time"
)

func fetch() error {
	return exec.Command("get-weather.sh").Run()
}

func show() string {
	cmd := exec.Command("show-weather.sh")
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

func Monitor() <-chan string {
	ch := make(chan string, 1)
	fetch()
	go monitor(ch)
	return ch
}

func monitor(ch chan<- string) {
	for {
		backoff := time.Second
		err := fetch()
		for err != nil {
			backoff *= 2
			if backoff > time.Minute / 4 {
				backoff = time.Minute / 4
			}
			time.Sleep(backoff)
			err = fetch()
		}
		ch <- show()
		time.Sleep(20 * time.Minute)
	}

}
