package main

import (
	"machine"
	"time"

	"github.com/HattoriHanzo031/gotto/board"
	"github.com/HattoriHanzo031/gotto/buzzer"
	"github.com/HattoriHanzo031/gotto/ninja"
	"github.com/HattoriHanzo031/gotto/servo"
	"tinygo.org/x/drivers/hcsr04"
	tgservo "tinygo.org/x/drivers/servo"
)

var (
	pwmFoot   = board.FootPWM
	pwmLeg    = board.LegPWM
	pwmBuzzer = board.BuzzerPWM

	rLegPin  = board.RLegPin  // Header 008 (P0.08)
	lLegPin  = board.LLegPin  // Header 017 (P0.17)
	rFootPin = board.RFootPin // Header 006 (P0.06)
	lFootPin = board.LFootPin // Header 020 (P0.20)

	buzzerPin = board.BuzzerPin // Header 031 (P0.31)
	led       = machine.LED
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// bitbangTone generates a square wave for passive piezos
func bitbangTone(pin machine.Pin, freqHz int, dur time.Duration) {
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

// probePinElectrical performs an electrical impedance & decay audit on a pin
func probePinElectrical(pin machine.Pin, name string) {
	println("  [ELECTRICAL AUDIT]", name, "...")

	// 1. Internal Pullup Test
	pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	time.Sleep(5 * time.Millisecond)
	pu := pin.Get()

	// 2. Internal Pulldown Test
	pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	time.Sleep(5 * time.Millisecond)
	pd := pin.Get()

	// 3. Floating High-Z Test
	pin.Configure(machine.PinConfig{Mode: machine.PinInput})
	time.Sleep(5 * time.Millisecond)
	flt := pin.Get()

	// 4. Decay / discharge test: charge pin HIGH, release to float, count decay cycles
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pin.High()
	time.Sleep(10 * time.Millisecond)
	pin.Configure(machine.PinConfig{Mode: machine.PinInput})
	decayCycles := 0
	for i := 0; i < 50000; i++ {
		if !pin.Get() {
			decayCycles = i
			break
		}
	}

	println("    Pull-up level:   ", pu)
	println("    Pull-down level: ", pd)
	println("    Floating level:  ", flt)
	println("    Discharge cycles:", decayCycles)
	if decayCycles < 100 {
		println("    RESULT: Physical load DETECTED (rapid discharge / circuit connected to GND)")
	} else {
		println("    RESULT: High-impedance / open line (pin holds charge)")
	}
}

func beep(bz *buzzer.Buzzer, count int) {
	for i := 0; i < count; i++ {
		led.Low() // LED on
		_ = bz.Tone(buzzer.Note{Period: buzzer.C5, Duration: 100 * time.Millisecond})
		led.High() // LED off
		time.Sleep(80 * time.Millisecond)
	}
}

func main() {
	time.Sleep(2 * time.Second)
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.High()

	println("\n=======================================================")
	println("  GOTTO HARDWARE DIAGNOSTIC & TELEMETRY SUITE          ")
	println("=======================================================")

	// 1. ENABLE VCC POWER RAIL
	println("\n[INIT] Enabling 3.3V VCC Rail (P0.13 HIGH)...")
	board.EnableVcc()
	time.Sleep(300 * time.Millisecond)
	println("  VCC rail enabled.")

	// 2. ELECTRICAL PROBE ON PIN 031 (Buzzer) & 113
	println("\n[INIT] Probing Electrical Connectivity on Header 031 (P0.31)...")
	probePinElectrical(buzzerPin, "Header 031 (Buzzer)")
	probePinElectrical(machine.P1_13, "Header 113 (Old Buzzer Pin)")

	// 3. SENSOR PROBE: HC-SR04 ULTRASONIC SENSOR
	println("\n[INIT] Probing Ultrasonic Sensor (Trig: 100 / Echo: 011)...")
	us := hcsr04.New(board.USTrigPin, board.USEchoPin)
	us.Configure()
	dist := us.ReadDistance()
	println("  Ultrasonic sensor read distance:", dist, "mm")

	// 4. CONFIGURE BUZZER ON HEADER 031
	println("\n[INIT] Configuring Buzzer on Header 031 (P0.31)...")
	bz := buzzer.New(buzzer.NewPwmChannel(pwmBuzzer, buzzerPin))
	if err := bz.Configure(); err != nil {
		println("  PWM0 configure error:", err.Error())
	} else {
		println("  Buzzer PWM0 configured successfully.")
	}

	// 5. CONFIGURE SERVOS (006, 008, 017, 020)
	println("\n[INIT] Configuring Servos on Headers 006, 008, 017, 020...")
	legArr := must(tgservo.NewArray(pwmLeg))
	footArr := must(tgservo.NewArray(pwmFoot))

	llServo := servo.New180(must(legArr.Add(lLegPin)), 450, 2550)
	rlServo := servo.New180(must(legArr.Add(rLegPin)), 450, 2550)
	lfServo := servo.New360(must(footArr.Add(lFootPin)), 450, 2550)
	rfServo := servo.New360(must(footArr.Add(rFootPin)), 450, 2550)
	println("  All 4 servos configured successfully.")

	// Initialize Ninja robot
	n := ninja.New(rlServo, llServo, rfServo, lfServo, bz)
	n.Trim(ninja.Trim{
		TiltAngle:         0,
		LeftStepDuration:  0,
		RightStepDuration: 150 * time.Millisecond,
		LfSpeed:           0,
		RfSpeed:           0,
		LlAngle:           20,
		RlAngle:           12,
	})
	_ = n.Mode(ninja.ModeWalk)

	cycle := 1
	for {
		println("\n-------------------------------------------------------")
		println(">>> STARTING TEST CYCLE #", cycle, "<<<")
		println("-------------------------------------------------------")

		// ----------------------------------------------------
		// PHASE 1: SENSOR TELEMETRY REPORT
		// ----------------------------------------------------
		dist = us.ReadDistance()
		println("\n[TELEMETRY] Live Ultrasonic Distance:", dist, "mm")

		// ----------------------------------------------------
		// PHASE 2: BUZZER TESTS ON HEADER 031
		// ----------------------------------------------------
		println("\n[PHASE 2] Exercising Buzzer on Header 031 (P0.31)...")

		// A: PWM0 Musical Chime
		println("  (2A) Playing PWM0 melody (C4 -> E4 -> G4 -> C5 -> E5)...")
		_ = bz.Tone(buzzer.Note{Period: buzzer.C4, Duration: 150 * time.Millisecond})
		_ = bz.Tone(buzzer.Note{Period: buzzer.E4, Duration: 150 * time.Millisecond})
		_ = bz.Tone(buzzer.Note{Period: buzzer.G4, Duration: 150 * time.Millisecond})
		_ = bz.Tone(buzzer.Note{Period: buzzer.C5, Duration: 200 * time.Millisecond})
		_ = bz.Tone(buzzer.Note{Period: buzzer.E5, Duration: 300 * time.Millisecond})
		time.Sleep(300 * time.Millisecond)

		// B: 2.7 kHz resonant frequency bit-bang
		println("  (2B) Playing 2.7 kHz resonant pulse (piezo resonance)...")
		buzzerPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		bitbangTone(buzzerPin, 2700, 300*time.Millisecond)
		time.Sleep(300 * time.Millisecond)

		// C: Digital steady HIGH pulse (active buzzer)
		println("  (2C) Pulsing digital HIGH (active buzzer)...")
		buzzerPin.High()
		time.Sleep(250 * time.Millisecond)
		buzzerPin.Low()
		time.Sleep(300 * time.Millisecond)

		// D: Rapid multi-tone chirp sequence
		println("  (2D) Playing rapid multi-tone chirp (1 kHz -> 2 kHz -> 4 kHz)...")
		bitbangTone(buzzerPin, 1000, 100*time.Millisecond)
		bitbangTone(buzzerPin, 2000, 100*time.Millisecond)
		bitbangTone(buzzerPin, 4000, 150*time.Millisecond)
		time.Sleep(400 * time.Millisecond)

		// Re-configure PWM0 for subsequent servo beeps
		_ = bz.Configure()

		// ----------------------------------------------------
		// PHASE 3: INDIVIDUAL SERVO TESTS
		// ----------------------------------------------------
		println("\n[PHASE 3] Testing Each Servo Individually...")

		// Motor 1: Right Foot on 006
		println("  -> Motor 1: Right Foot on Header 006 (1 beep, continuous spin forward/reverse)...")
		beep(bz, 1)
		time.Sleep(200 * time.Millisecond)
		_ = rfServo.SetSpeed(35)
		time.Sleep(600 * time.Millisecond)
		_ = rfServo.SetSpeed(-35)
		time.Sleep(600 * time.Millisecond)
		_ = rfServo.SetSpeed(0)
		time.Sleep(400 * time.Millisecond)

		// Motor 2: Right Leg on 008
		println("  -> Motor 2: Right Leg on Header 008 (2 beeps, angle sweep 90->65->115->90)...")
		beep(bz, 2)
		time.Sleep(200 * time.Millisecond)
		for _, angle := range []int{90, 65, 115, 90} {
			_ = rlServo.SetAngle(angle)
			time.Sleep(400 * time.Millisecond)
		}
		time.Sleep(400 * time.Millisecond)

		// Motor 3: Left Leg on 017
		println("  -> Motor 3: Left Leg on Header 017 (3 beeps, angle sweep 90->65->115->90)...")
		beep(bz, 3)
		time.Sleep(200 * time.Millisecond)
		for _, angle := range []int{90, 65, 115, 90} {
			_ = llServo.SetAngle(angle)
			time.Sleep(400 * time.Millisecond)
		}
		time.Sleep(400 * time.Millisecond)

		// Motor 4: Left Foot on 020
		println("  -> Motor 4: Left Foot on Header 020 (4 beeps, continuous spin forward/reverse)...")
		beep(bz, 4)
		time.Sleep(200 * time.Millisecond)
		_ = lfServo.SetSpeed(35)
		time.Sleep(600 * time.Millisecond)
		_ = lfServo.SetSpeed(-35)
		time.Sleep(600 * time.Millisecond)
		_ = lfServo.SetSpeed(0)
		time.Sleep(400 * time.Millisecond)

		// ----------------------------------------------------
		// PHASE 4: COORDINATED BIPED LOCOMOTION
		// ----------------------------------------------------
		println("\n[PHASE 4] Coordinated Biped Locomotion...")
		println("  -> Centering to Home position...")
		_ = n.Home()
		time.Sleep(500 * time.Millisecond)

		println("  -> Walking 2 steps forward...")
		_ = n.Walk(2)
		time.Sleep(500 * time.Millisecond)

		println("  -> Performing Happy Wiggle gesture...")
		_ = n.HappyWiggle()
		time.Sleep(500 * time.Millisecond)

		println("  -> Walking 2 steps backward (Moonwalk)...")
		_ = n.Moonwalk(2)
		time.Sleep(500 * time.Millisecond)

		println("\n[CYCLE #", cycle, " COMPLETE] Resting 6 seconds before next cycle...")
		time.Sleep(6 * time.Second)
		cycle++
	}
}
