package weather

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	WOEID_BOSTON = "2367105"

	YWeatherEndpoint = "https://weather.yahooapis.com/forecastrss?w="
)

func fetch() error {
	resp, err := http.Get(YWeatherEndpoint + WOEID_BOSTON)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	file, err := os.Create(os.Getenv("HOME") + "/.cache/weather")
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	return err
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
