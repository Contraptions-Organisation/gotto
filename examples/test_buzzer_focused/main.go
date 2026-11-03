package main

import (
	"machine"
	"time"

	"github.com/HattoriHanzo031/gotto/board"
)

func toneActiveHigh(pin machine.Pin, freq int, dur time.Duration) {
	halfPeriod := time.Duration(1000000/(2*freq)) * time.Microsecond
	cycles := int(dur / (2 * halfPeriod))
	for i := 0; i < cycles; i++ {
		pin.High()
		time.Sleep(halfPeriod)
		pin.Low()
		time.Sleep(halfPeriod)
	}
}

func toneActiveLow(pin machine.Pin, freq int, dur time.Duration) {
	halfPeriod := time.Duration(1000000/(2*freq)) * time.Microsecond
	cycles := int(dur / (2 * halfPeriod))
	for i := 0; i < cycles; i++ {
		pin.Low()
		time.Sleep(halfPeriod)
		pin.High()
		time.Sleep(halfPeriod)
	}
}

func testPin(name string, pin machine.Pin, led machine.Pin) {
	println("\n>>> PROBING PIN:", name, "<<<")
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// 1. ACTIVE LOW (Low level triggers many Arduino buzzer modules with PNP/NPN)
	println("  [1] Active-LOW continuous (500ms)...")
	led.Low() // LED on
	pin.Low()
	time.Sleep(500 * time.Millisecond)
	pin.High()
	led.High() // LED off
	time.Sleep(200 * time.Millisecond)

	// 2. ACTIVE HIGH (High level triggers direct buzzers)
	println("  [2] Active-HIGH continuous (500ms)...")
	led.Low()
	pin.High()
	time.Sleep(500 * time.Millisecond)
	pin.Low()
	led.High()
	time.Sleep(200 * time.Millisecond)

	// 3. 2.7 kHz resonant frequency (Standard piezo buzzer pitch)
	println("  [3] 2.7 kHz square wave (700ms)...")
	led.Low()
	toneActiveHigh(pin, 2700, 700*time.Millisecond)
	led.High()
	time.Sleep(200 * time.Millisecond)

	// 4. 1.0 kHz square wave
	println("  [4] 1.0 kHz square wave (700ms)...")
	led.Low()
	toneActiveHigh(pin, 1000, 700*time.Millisecond)
	led.High()
	time.Sleep(200 * time.Millisecond)

	// 5. Active-LOW 2.7 kHz wave
	println("  [5] Active-LOW 2.7 kHz square wave (700ms)...")
	led.Low()
	toneActiveLow(pin, 2700, 700*time.Millisecond)
	led.High()
	pin.High() // leave high (idle off for active-low)
	time.Sleep(500 * time.Millisecond)
}

func main() {
	time.Sleep(1 * time.Second)

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	println("==================================================")
	println("     FOCUSED BUZZER DIAGNOSTIC + LED SYNC         ")
	println("==================================================")
	println("Watch the Nice!Nano onboard LED - it will light up")
	println("during each sound attempt to prove MCU execution.")

	board.EnableVcc()
	time.Sleep(200 * time.Millisecond)

	primaryPins := []struct {
		name string
		pin  machine.Pin
	}{
		{"P0.31 (D031 - Pin 5 / CirkitDesigner schematic)", machine.P0_31},
		{"P1.13 (D113 - Pin 9 / board.go)", machine.P1_13},
		{"P0.29 (D029 - Header 029)", machine.P0_29},
		{"P0.02 (D002 - Header 002)", machine.P0_02},
		{"P0.22 (D022 - Header 022)", machine.P0_22},
		{"P0.24 (D024 - Header 024)", machine.P0_24},
	}

	cycle := 1
	for {
		println("\n--- STARTING FOCUSED CYCLE", cycle, "---")
		for _, p := range primaryPins {
			testPin(p.name, p.pin, led)
			time.Sleep(1 * time.Second)
		}
		cycle++
		time.Sleep(2 * time.Second)
	}
}
