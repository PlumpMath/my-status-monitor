package weather

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type WeatherConditionsInfo struct {
	WeatherType string `json:"weather-type"`
	Coverage    string `json:"coverage"`
	Intensity   string `json:"intensity"`
}

type WeatherInfo struct {
	Temp       float64                `json:"temp"`
	Units      string                 `json:"units"`
	Conditions *WeatherConditionsInfo `json:"conditions"`
}

func (i *WeatherInfo) Show() string {
	var Conditions string
	if i.Conditions != nil {
		Conditions = fmt.Sprintf(
			", %s %s %s",
			i.Conditions.Coverage,
			i.Conditions.Intensity,
			i.Conditions.WeatherType,
		)
	}
	return fmt.Sprintf("%v %s%s", i.Temp, i.Units, Conditions)
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
