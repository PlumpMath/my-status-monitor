package battery

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type State struct {
	// true iff the battery is being charged.
	Charging bool
	// The percentage of the battery's maximum energy remaining.
	PercentRemaining int
	// Time until battery is exhausted. Only meaningful if Charging is false.
	Hours, Minutes, Seconds int
}

type Battery struct {
	Path  string
	State *State
}

func (s *State) String() string {
	if s.Charging {
		return fmt.Sprintf("Battery, charging, %d%%", s.PercentRemaining)
	} else {
		return fmt.Sprintf("On battery, %d:%02d remaining (%d%%)",
			s.Hours, s.Minutes, s.PercentRemaining)
	}
}

func (b *Battery) Poll() {
	var err error

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
		return
	}

	result := &State{}

	// work out how much time we have left
	result.Hours = energy / power
	energy %= power
	result.Minutes = (60 * energy) / power
	energy = (60 * energy) % power
	result.Seconds = (60 * energy) / power

	result.PercentRemaining = capacity

	result.Charging = (text["status"][0] == 'C')
	b.State = result
}

func (b *Battery) Monitor() {
	if !b.exists() {
		return
	}
	for {
		b.Poll()
		time.Sleep(time.Second)
	}
}

func (b *Battery) NotifyDaemon() {
	if !b.exists() {
		return
	}
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
		case state.Minutes < 5:
			notifySend("critical", "Very low battery", state.String())
		case state.Minutes < 20:
			notifySend("normal", "Low battery", state.String())
		}

		// Check for AC adapter events.
		switch {
		case oldstate == nil || state == nil || oldstate.Charging == state.Charging:
			// The AC adapter hasn't changed state since we last measured it.
		case state.Charging:
			notifySend("normal", "AC adapter connected", state.String())
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

func (b *Battery) exists() bool {
	_, err := os.Stat(b.Path)
	return err == nil
}
