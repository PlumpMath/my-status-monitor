package weather

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type WeatherInfo struct {
	CurrentTemp float64 `json:"current_temp"`
	Units       string  `json:"units"`
}

func (i *WeatherInfo) Show() string {
	return fmt.Sprintf("%v %s", i.CurrentTemp, i.Units)
}

func fetch() (*WeatherInfo, error) {
	bytes, err := exec.Command("noaa.py").Output()
	if err != nil {
		return nil, err
	}
	ret := &WeatherInfo{}
	err = json.Unmarshal(bytes, ret)
	return ret, err
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
		info, err := fetch()
		for err != nil {
			backoff *= 2
			if backoff > time.Minute/4 {
				backoff = time.Minute / 4
			}
			time.Sleep(backoff)
			info, err = fetch()
		}
		ch <- info.Show()
		time.Sleep(20 * time.Minute)
	}

}
