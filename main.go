package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	timeFormat = "Mon Jan 2 3:04 PM"
	batteryPath = "/sys/class/power_supply/BAT0/"
)

var (
	deviceNotReady = errors.New("Device is not ready for polling")
)

var (
	battInfo *BatteryInfo
)

type BatteryInfo struct {
	hour, min, sec int // time remaining
	capacity int // percent charged
	charging bool
}

func (b *BatteryInfo) String() string {
	if b.charging {
		return fmt.Sprintf("Battery charging, %d%%", b.capacity)
	} else {
		return fmt.Sprintf("On battery, %d:%02d:%02d remaining (%d%%)",
			b.hour, b.min, b.sec, b.capacity)
	}
}

func emit(str string) {
	cmd := exec.Command("xsetroot", "-name", str)
	cmd.Run()
}

func haveBattery() bool {
	_, err := os.Stat(batteryPath)
	return err == nil
}

func slurpFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	return string(data), err
}

func pollBattery() (*BatteryInfo, error) {
	var err error
	text := make(map[string] string)
	for _, f := range []string{"status", "capacity", "power_now", "energy_now" } {
		text[f], err = slurpFile(batteryPath + f)
		if err != nil {
			return nil, err
		}
		text[f] = strings.TrimSpace(text[f])
	}

	nums := make(map[string] int)
	for _, f := range []string{"capacity", "power_now", "energy_now"} {
		nums[f], err = strconv.Atoi(text[f])
		if err != nil {
			fmt.Printf("%#v", text[f])
			return nil, err
		}
	}

	capacity, power, energy := nums["capacity"], nums["power_now"], nums["energy_now"]

	if power == 0 {
		// We should wait for the device to settle.
		return nil, deviceNotReady
	}

	result := &BatteryInfo{}

	// work out how much time we have left
	result.hour = energy/power
	energy %= power
	result.min = (60*energy) / power
	energy = (60*energy) % power
	result.sec = (60*energy) / power

	result.capacity = capacity

	result.charging = (text["status"][0] == 'C')
	return result, nil
}

func main() {
	go func() {
		if !haveBattery() {
			return
		}
		for {
			info, err := pollBattery()
			if err == nil {
				battInfo = info
			}
			time.Sleep(time.Second)
		}
	}()

	for {
		t := time.Now()
		batt := battInfo
		if batt != nil {
			emit(fmt.Sprintf("%v | %s", batt, t.Format(timeFormat)))
		} else {
			emit(t.Format(timeFormat))
		}
		time.Sleep(time.Second)
	}
}
