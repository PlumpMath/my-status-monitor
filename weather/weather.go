package weather

import (
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"fmt"
)

var (
	httpNotOK = errors.New("HTTP response was not 200 OK")
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
	if resp.StatusCode != 200 {
		return httpNotOK
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
	go monitor(ch)
	return ch
}

func monitor(ch chan<- string) {
	for {
		backoff := time.Second
		err := fetch()
		for err != nil {
			fmt.Println(err)
			backoff *= 2
			if backoff > time.Minute / 2 {
				backoff = time.Minute / 2
			}
			fmt.Println("backoff = ", backoff)
			time.Sleep(backoff)
			err = fetch()
		}
		ch <- show()
		time.Sleep(20 * time.Minute)
	}

}
