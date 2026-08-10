// Command pinscan drives a servo-style PWM signal to every broken-out GPIO
// pin on the nice!nano / Super Mini header, one at a time, so you can watch
// (or listen to) the robot and figure out which physical pin actually moves
// which servo when the wiring is unknown or in doubt.
//
// Skips D004 (battery voltage ADC input) and D013 (VCC power switch) since
// driving those is either meaningless or can cut power to the board.
package main

import (
	"machine"
	"time"

	tgservo "tinygo.org/x/drivers/servo"
)

var candidatePins = []struct {
	name string
	pin  machine.Pin
}{
	{"D002 (P0.02, AIN0)", machine.D002},
	{"D006 (P0.06)", machine.D006},
	{"D008 (P0.08)", machine.D008},
	{"D009 (P0.09, NFC1)", machine.D009},
	{"D010 (P0.10, NFC2)", machine.D010},
	{"D011 (P0.11)", machine.D011},
	{"D012 (P0.12)", machine.D012},
	{"D017 (P0.17)", machine.D017},
	{"D020 (P0.20)", machine.D020},
	{"D022 (P0.22)", machine.D022},
	{"D024 (P0.24)", machine.D024},
	{"D026 (P0.26)", machine.D026},
	{"D029 (P0.29, AIN5)", machine.D029},
	{"D031 (P0.31, AIN7)", machine.D031},
	{"D100 (P1.00)", machine.D100},
	{"D101 (P1.01)", machine.D101},
	{"D102 (P1.02)", machine.D102},
	{"D104 (P1.04)", machine.D104},
	{"D106 (P1.06)", machine.D106},
	{"D107 (P1.07)", machine.D107},
	{"D111 (P1.11)", machine.D111},
	{"D113 (P1.13)", machine.D113},
	{"D115 (P1.15)", machine.D115},
}

var pwmPeripherals = []struct {
	name string
	pwm  *machine.PWM
}{
	{"PWM0", machine.PWM0},
	{"PWM1", machine.PWM1},
	{"PWM2", machine.PWM2},
	{"PWM3", machine.PWM3},
}

func main() {
	machine.InitSerial()
	time.Sleep(3 * time.Second)

	println("GOtto pin scan starting.")
	println("Watch/listen to the robot. Each pin gets ~2.5s: full one way, full the other, then center.")
	println("Skipping D004 (battery ADC) and D013 (VCC power switch) for safety.")
	println()

	for _, p := range candidatePins {
		driven := false
		for _, dev := range pwmPeripherals {
			arr, err := tgservo.NewArray(dev.pwm)
			if err != nil {
				continue
			}
			s, err := arr.Add(p.pin)
			if err != nil {
				continue
			}
			println("-->", p.name, "via", dev.name)
			s.SetMicroseconds(1000)
			time.Sleep(1 * time.Second)
			s.SetMicroseconds(2000)
			time.Sleep(1 * time.Second)
			s.SetMicroseconds(1500)
			time.Sleep(500 * time.Millisecond)
			driven = true
			break
		}
		if !driven {
			println("    ", p.name, "- no PWM peripheral could claim this pin, skipped")
		}
	}

	println()
	println("Pin scan complete. Restart the board to run it again.")
	for {
		time.Sleep(time.Hour)
	}
}
