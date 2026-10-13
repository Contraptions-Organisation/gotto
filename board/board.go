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

	BuzzerPin = machine.P0_31

	ButtonPin = machine.P1_04

	// VccPin is the SuperMini/nice!nano switched VCC rail (P0.13). Driving
	// it high enables the onboard 3.3V regulator (powering the shield and servos);
	// driving it low cuts power to the rail.
	VccPin = machine.P0_13
)

// EnableVcc drives the switched VCC rail high, powering on anything wired to
// it (e.g. servos, buzzer, sensors on the shield). Call this at startup.
func EnableVcc() {
	VccPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	VccPin.High()
}

// DisableVcc drives the switched VCC rail low, cutting power to save energy.
func DisableVcc() {
	VccPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	VccPin.Low()
}
