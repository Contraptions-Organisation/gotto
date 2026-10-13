package main

import (
	"machine"
	"time"

	"github.com/HattoriHanzo031/gotto/buzzer"
)

var (
	buzzerPin = machine.P1_13 // Header 113
	pwm       = machine.PWM0
	led       = machine.LED
)

// busyWaitTone generates a precise square wave using time.Now() busy wait
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

	println("==================================================")
	println("   CONTINUOUS BUZZER TEST FOR PIN 113 (P1.13)     ")
	println("==================================================")

	cycle := 1
	for {
		println("\n>>> [CYCLE", cycle, "] TESTING PIN 113 <<<")

		// ----------------------------------------------------
		// 1. HARDWARE PWM MODE (TinyGo PWM0)
		// ----------------------------------------------------
		println("  (A) Hardware PWM0: A4 (440 Hz)...")
		led.Low() // LED on
		bz := buzzer.New(buzzer.NewPwmChannel(pwm, buzzerPin))
		if err := bz.Configure(); err == nil {
			_ = bz.Tone(buzzer.Note{Period: buzzer.A4, Duration: 1 * time.Second})
			time.Sleep(100 * time.Millisecond)

			println("  (B) Hardware PWM0: C5 (523 Hz)...")
			_ = bz.Tone(buzzer.Note{Period: buzzer.C5, Duration: 1 * time.Second})
			time.Sleep(100 * time.Millisecond)

			println("  (C) Hardware PWM0: C6 (1046 Hz)...")
			_ = bz.Tone(buzzer.Note{Period: buzzer.C6, Duration: 1 * time.Second})
		} else {
			println("  PWM0 error:", err.Error())
		}
		led.High() // LED off
		time.Sleep(300 * time.Millisecond)

		// ----------------------------------------------------
		// 2. ACTIVE BUZZER TEST: STEADY DIGITAL HIGH
		// ----------------------------------------------------
		println("  (D) Digital steady HIGH (Active Buzzer check, 1s)...")
		led.Low()
		buzzerPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		buzzerPin.High()
		time.Sleep(1 * time.Second)
		buzzerPin.Low()
		led.High()
		time.Sleep(300 * time.Millisecond)

		// ----------------------------------------------------
		// 3. ACTIVE BUZZER TEST: STEADY DIGITAL LOW
		// ----------------------------------------------------
		println("  (E) Digital steady LOW (Active-low buzzer check, 1s)...")
		led.Low()
		buzzerPin.Low()
		time.Sleep(1 * time.Second)
		buzzerPin.High()
		led.High()
		time.Sleep(300 * time.Millisecond)

		// ----------------------------------------------------
		// 4. PRECISE BUSY-WAIT BIT-BANG: 2.7 kHz resonant tone
		// ----------------------------------------------------
		println("  (F) Precise 2.7 kHz resonant square wave (1s)...")
		led.Low()
		busyWaitTone(buzzerPin, 2700, 1*time.Second)
		led.High()
		buzzerPin.Low()
		time.Sleep(300 * time.Millisecond)

		// ----------------------------------------------------
		// 5. PRECISE BUSY-WAIT BIT-BANG: 1.0 kHz square wave
		// ----------------------------------------------------
		println("  (G) Precise 1.0 kHz square wave (1s)...")
		led.Low()
		busyWaitTone(buzzerPin, 1000, 1*time.Second)
		led.High()
		buzzerPin.Low()
		time.Sleep(500 * time.Millisecond)

		cycle++
	}
}
