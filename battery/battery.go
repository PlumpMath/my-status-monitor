package battery

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

var (
	deviceBusy = errors.New("battery updating; can't poll.")
)

type state struct {
	// true iff the battery is being charged.
	Charging bool
	// The percentage of the battery's maximum energy remaining.
	PercentRemaining int
	// Time until battery is exhausted. Only meaningful if Charging is false.
	Hours, Minutes, Seconds int
}

type battery struct {
	Path  string
	State *state
}

func Monitor(filename string) <-chan string {
	ch := make(chan string, 1)
	b := &battery{filename, nil}
	go b.Monitor(ch)
	go b.NotifyDaemon()
	return ch
}

func (s *state) String() string {
	if s.Charging {
		return fmt.Sprintf("Battery charging, %d%%", s.PercentRemaining)
	} else {
		return fmt.Sprintf("On battery, %d:%02d remaining (%d%%)",
			s.Hours, s.Minutes, s.PercentRemaining)
	}
}

func (b *battery) Poll() (err error) {

	// Read in each of the files we care about, and store their contents in a map.
	text := make(map[string]string)
	for _, f := range []string{"status", "capacity", "power_now", "energy_now"} {
		text[f], err = slurpFile(b.Path + f)
		if err != nil {
			// If we're unsuccessful reading any of the files, stop now.
			// We don't want to update the state until we can get a
			// meaningful reading.
			return
		}
	}

	// Most of the files contain numbers; convert them into actual
	// integers so we can work with them.
	nums := make(map[string]int)
	for _, f := range []string{"capacity", "power_now", "energy_now"} {
		// The call to TrimSpace is important; the files end with a
		// newline, so without this Atoi will fail.
		nums[f], err = strconv.Atoi(strings.TrimSpace(text[f]))
		if err != nil {
			// Again, stop now if anything fails.
			return
		}
	}

	// assigning these to variables will make the code below less verbose
	capacity, power, energy := nums["capacity"], nums["power_now"], nums["energy_now"]

	if power == 0 {
		// If power is zero, it is usually an indication that the
		// device has been recently plugged in or unplugged, and
		// hasn't settled down yet. Stop here; we want to wait until
		// we can get an accurate reading.
		return deviceBusy
	}

	result := &state{}

	// work out how much time we have left
	result.Hours = energy / power
	energy %= power
	result.Minutes = (60 * energy) / power
	energy = (60 * energy) % power
	result.Seconds = (60 * energy) / power

	result.PercentRemaining = capacity

	result.Charging = (text["status"][0] == 'C')
	b.State = result
	return nil
}

func (b *battery) Monitor(ch chan<- string) {
	if !b.exists() {
		return
	}
	for {
		if b.Poll() == nil {
			ch <-b.State.String()
		}
		time.Sleep(time.Second)
	}
}

func (b *battery) NotifyDaemon() {
	if !b.exists() {
		return
	}

	// These variables record whether or not the battery has reached low
	// or very low states since the last time it was unplugged. This lets
	// us avoid getting multiple messages about those events, which
	// happens because the estimated time remaining tends to fluctuate.
	wasLow := false
	wasVeryLow := false

	oldstate := b.State
	for {
		state := b.State

		// Check for low battery status
		switch {
		case state == nil:
			// We don't have any data on the battery right now.
		case state.Hours > 0 || state.Charging:
			// Either the battery is not low, or or it's charging.
			// No need to bother the user.
		case state.Minutes < 5 && !wasVeryLow:
			notifySend("critical", "Very low battery", state.String())
			wasVeryLow = true
		case state.Minutes < 20 && !wasLow:
			notifySend("normal", "Low battery", state.String())
			wasLow = true
		}

		// Check for AC adapter events.
		switch {
		case oldstate == nil || state == nil || oldstate.Charging == state.Charging:
			// The AC adapter hasn't changed state since we last measured it.
		case state.Charging:
			notifySend("normal", "AC adapter connected", state.String())
			wasVeryLow = false
			wasLow = false
		case !state.Charging:
			notifySend("normal", "AC adapter disconnected", state.String())
		}

		// Save the current state for later comparison.
		oldstate = state

		time.Sleep(time.Second)
	}
}

func notifySend(urgency, summary, body string) {
	cmd := exec.Command("notify-send", "-u", urgency, summary, body)
	cmd.Run()
}

// This is a utility function which grabs the entire contents of a file by
// name. We use this in the Poll method.
func slurpFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	return string(data), err
}

func (b *battery) exists() bool {
	_, err := os.Stat(b.Path)
	return err == nil
}
