// Package board defines the pin and PWM mapping for the physical nRF52840-based
// controller board used by the OttoNinja robot (nice!nano-style "Super Mini" pinout).
package board

import "machine"

// PWM peripherals used by the controller board.
var (
	FootPWM   = machine.PWM2
	LegPWM    = machine.PWM1
	BuzzerPWM = machine.PWM0
)

// GPIO pins used by the controller board.
const (
	RLegPin  = machine.P0_08
	LLegPin  = machine.P0_17
	RFootPin = machine.P0_06
	LFootPin = machine.P0_20

	USTrigPin = machine.P1_00
	USEchoPin = machine.P0_11

	BuzzerPin = machine.P1_13

	ButtonPin = machine.P1_04

	// VccPin is the nice!nano/Super Mini switched VCC rail (P0.13). Driving
	// it high cuts power to the rail (saving battery when idle); driving it
	// low enables it. Servos are powered from this switched rail, so it must
	// be enabled before configuring or moving any servo.
	VccPin = machine.P0_13
)

// EnableVcc drives the switched VCC rail low, powering on anything wired to
// it (e.g. servos). Call this before configuring servos.
func EnableVcc() {
	VccPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	VccPin.Low()
}
