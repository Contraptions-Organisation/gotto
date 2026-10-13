package main

import (
	"machine"
	"time"

	"github.com/HattoriHanzo031/gotto/board"
)

var candidatePins = []struct {
	name string
	pin  machine.Pin
}{
	{"P1.13 (D113 - Default board.BuzzerPin)", machine.P1_13},
	{"P0.31 (D031 - Alternate old buzzerPin)", machine.P0_31},
	{"P0.29 (D029 - Header 029)", machine.P0_29},
	{"P0.02 (D002 - Header 002)", machine.P0_02},
	{"P1.15 (D115 - Header 115)", machine.P1_15},
	{"P1.11 (D111 - Header 111)", machine.P1_11},
	{"P0.22 (D022 - Header 022)", machine.P0_22},
	{"P0.24 (D024 - Header 024)", machine.P0_24},
	{"P1.00 (D100 - Header 100)", machine.P1_00},
	{"P0.11 (D011 - Header 011)", machine.P0_11},
	{"P1.04 (D104 - Header 104)", machine.P1_04},
	{"P1.06 (D106 - Header 106)", machine.P1_06},
	{"P1.01 (D101 - Header 101)", machine.P1_01},
	{"P1.02 (D102 - Header 102)", machine.P1_02},
	{"P1.07 (D107 - Header 107)", machine.P1_07},
	{"P0.10 (D010 - Header 010)", machine.P0_10},
	{"P0.09 (D009 - Header 009)", machine.P0_09},
}

// playTone generates a square wave on the given pin at the requested frequency.
func playTone(pin machine.Pin, freqHz int, duration time.Duration) {
	halfPeriod := time.Duration(1000000/(2*freqHz)) * time.Microsecond
	cycles := int(duration / (2 * halfPeriod))
	for i := 0; i < cycles; i++ {
		pin.High()
		time.Sleep(halfPeriod)
		pin.Low()
		time.Sleep(halfPeriod)
	}
}

func main() {
	time.Sleep(2 * time.Second)

	println("==================================================")
	println("            GOTTO BUZZER SCANNER / FINDER         ")
	println("==================================================")
	println("Enabling VCC rail (P0.13 LOW) so shield has power...")
	board.EnableVcc()
	time.Sleep(500 * time.Millisecond)

	cycle := 1
	for {
		println("\n--------------------------------------------------")
		println(">>> STARTING BUZZER SCAN CYCLE", cycle, "<<<")
		println("Listen closely. Each pin will be tested for ~1.5s.")
		println("--------------------------------------------------")

		for i, cp := range candidatePins {
			println("\n[", i+1, "/", len(candidatePins), "] TESTING PIN:", cp.name)

			// Configure as Output
			cp.pin.Configure(machine.PinConfig{Mode: machine.PinOutput})

			// 1. Passive Buzzer: 2.7 kHz resonant tone (common piezo resonance)
			println("    -> Pattern 1: 2.7 kHz tone (400ms)")
			playTone(cp.pin, 2700, 400*time.Millisecond)
			time.Sleep(100 * time.Millisecond)

			// 2. Passive Buzzer: 1 kHz tone (standard pitch)
			println("    -> Pattern 2: 1.0 kHz tone (400ms)")
			playTone(cp.pin, 1000, 400*time.Millisecond)
			time.Sleep(100 * time.Millisecond)

			// 3. Active Buzzer: Digital HIGH pulses
			println("    -> Pattern 3: Active HIGH pulses (300ms)")
			for p := 0; p < 3; p++ {
				cp.pin.High()
				time.Sleep(80 * time.Millisecond)
				cp.pin.Low()
				time.Sleep(40 * time.Millisecond)
			}

			// Ensure pin is low after test
			cp.pin.Low()

			// Pause between pins
			time.Sleep(800 * time.Millisecond)
		}

		println("\n>>> SCAN CYCLE", cycle, "COMPLETE. REPEATING IN 3 SECONDS... <<<")
		time.Sleep(3 * time.Second)
		cycle++
	}
}
