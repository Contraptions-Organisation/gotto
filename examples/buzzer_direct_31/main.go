package main

import (
	"machine"
	"time"

	"github.com/HattoriHanzo031/gotto/buzzer"
)

var (
	buzzerPin = machine.P0_31
	pwm       = machine.PWM0
	led       = machine.LED
)

// busyWaitTone generates a precise square wave on the pin
func busyWaitTone(pin machine.Pin, freqHz int, dur time.Duration) {
	halfPeriodUs := int64(1000000 / (2 * freqHz))
	end := time.Now().Add(dur)
	for time.Now().Before(end) {
		pin.High()
		t0 := time.Now()
		for time.Since(t0).Microseconds() < halfPeriodUs {
		}
		pin.Low()
		t1 := time.Now()
		for time.Since(t1).Microseconds() < halfPeriodUs {
		}
	}
}

func main() {
	time.Sleep(1 * time.Second)

	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.High() // LED off

	println("\n=======================================================")
	println("  DIRECT BUZZER TEST FOR PIN 0.31 (P0.31)              ")
	println("  Wiring: One leg to P0.31, other leg to GND           ")
	println("=======================================================")

	bz := buzzer.New(buzzer.NewPwmChannel(pwm, buzzerPin))
	pwmConfigured := false
	if err := bz.Configure(); err == nil {
		pwmConfigured = true
		println("  Hardware PWM0 initialized on P0.31.")
	} else {
		println("  PWM0 configure note:", err.Error())
	}

	cycle := 1
	for {
		println("\n>>> [CYCLE", cycle, "] RUNNING ALL SOUND MODES ON PIN P0.31 <<<")

		// ----------------------------------------------------
		// 1. PIEZO RESONANCE: 2.7 kHz (standard piezo loud band)
		// ----------------------------------------------------
		println("  (1) Passive Piezo: 2.7 kHz resonant wave (500ms)...")
		buzzerPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		led.Low() // LED ON
		busyWaitTone(buzzerPin, 2700, 500*time.Millisecond)
		led.High() // LED OFF
		buzzerPin.Low()
		time.Sleep(200 * time.Millisecond)

		// ----------------------------------------------------
		// 2. AUDIBLE TONE: 1.0 kHz
		// ----------------------------------------------------
		println("  (2) Passive Piezo: 1.0 kHz square wave (500ms)...")
		led.Low()
		busyWaitTone(buzzerPin, 1000, 500*time.Millisecond)
		led.High()
		buzzerPin.Low()
		time.Sleep(200 * time.Millisecond)

		// ----------------------------------------------------
		// 3. AUDIBLE TONE: 440 Hz (A4)
		// ----------------------------------------------------
		println("  (3) Passive Piezo: 440 Hz (A4) square wave (500ms)...")
		led.Low()
		busyWaitTone(buzzerPin, 440, 500*time.Millisecond)
		led.High()
		buzzerPin.Low()
		time.Sleep(200 * time.Millisecond)

		// ----------------------------------------------------
		// 4. ACTIVE BUZZER: Steady DC HIGH (3.3V)
		// ----------------------------------------------------
		println("  (4) Active Buzzer check: Steady DC HIGH 3.3V (800ms)...")
		led.Low()
		buzzerPin.High()
		time.Sleep(800 * time.Millisecond)
		buzzerPin.Low()
		led.High()
		time.Sleep(200 * time.Millisecond)

		// ----------------------------------------------------
		// 5. RAPID CHIRP: 1.5 kHz -> 2.5 kHz -> 3.5 kHz -> 4.5 kHz
		// ----------------------------------------------------
		println("  (5) Multi-frequency chirp sequence (400ms)...")
		led.Low()
		busyWaitTone(buzzerPin, 1500, 100*time.Millisecond)
		busyWaitTone(buzzerPin, 2500, 100*time.Millisecond)
		busyWaitTone(buzzerPin, 3500, 100*time.Millisecond)
		busyWaitTone(buzzerPin, 4500, 100*time.Millisecond)
		led.High()
		buzzerPin.Low()
		time.Sleep(200 * time.Millisecond)

		// ----------------------------------------------------
		// 6. HARDWARE PWM0: Musical Melody (Imperial March)
		// ----------------------------------------------------
		if pwmConfigured {
			println("  (6) Hardware PWM0: Imperial March melody...")
			_ = bz.Configure()
			led.Low()
			_ = bz.Tone(buzzer.Note{Period: buzzer.A4, Duration: 350 * time.Millisecond})
			time.Sleep(100 * time.Millisecond)
			_ = bz.Tone(buzzer.Note{Period: buzzer.A4, Duration: 350 * time.Millisecond})
			time.Sleep(100 * time.Millisecond)
			_ = bz.Tone(buzzer.Note{Period: buzzer.A4, Duration: 350 * time.Millisecond})
			time.Sleep(100 * time.Millisecond)
			_ = bz.Tone(buzzer.Note{Period: buzzer.F4, Duration: 250 * time.Millisecond})
			time.Sleep(50 * time.Millisecond)
			_ = bz.Tone(buzzer.Note{Period: buzzer.C5, Duration: 150 * time.Millisecond})
			time.Sleep(50 * time.Millisecond)
			_ = bz.Tone(buzzer.Note{Period: buzzer.A4, Duration: 400 * time.Millisecond})
			led.High()
		}

		println("\n[CYCLE", cycle, "COMPLETE] Resting 1.5 seconds...")
		time.Sleep(1500 * time.Millisecond)
		cycle++
	}
}
