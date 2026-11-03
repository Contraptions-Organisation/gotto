package main

import (
	"machine"
	"time"

	"github.com/HattoriHanzo031/gotto/board"
	"github.com/HattoriHanzo031/gotto/buzzer"
	"github.com/HattoriHanzo031/gotto/ninja"
	"github.com/HattoriHanzo031/gotto/servo"
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

// bitbangTone generates a resonant tone
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

func beep(bz *buzzer.Buzzer, count int) {
	for i := 0; i < count; i++ {
		led.Low() // LED on
		_ = bz.Tone(buzzer.Note{Period: buzzer.C5, Duration: 120 * time.Millisecond})
		led.High() // LED off
		time.Sleep(80 * time.Millisecond)
	}
}

func main() {
	machine.InitSerial()
	time.Sleep(2 * time.Second)

	board.EnableVcc()
	time.Sleep(200 * time.Millisecond)

	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.High()

	println("\n=======================================================")
	println("  GOTTO BIPED WALKING & BUZZER FIRMWARE                ")
	println("=======================================================")

	// 1. CONFIGURE BUZZER ON PIN 031
	println("\n[INIT] Configuring Buzzer on Header 031 (P0.31)...")
	bz := buzzer.New(buzzer.NewPwmChannel(pwmBuzzer, buzzerPin))
	if err := bz.Configure(); err != nil {
		println("  PWM0 error:", err.Error())
	} else {
		println("  Buzzer initialized. Playing startup fanfare...")
		_ = bz.Tone(buzzer.Note{Period: buzzer.C4, Duration: 150 * time.Millisecond})
		_ = bz.Tone(buzzer.Note{Period: buzzer.E4, Duration: 150 * time.Millisecond})
		_ = bz.Tone(buzzer.Note{Period: buzzer.G4, Duration: 150 * time.Millisecond})
		_ = bz.Tone(buzzer.Note{Period: buzzer.C5, Duration: 250 * time.Millisecond})
	}

	// 2. CONFIGURE SERVOS (006, 008, 017, 020)
	println("\n[INIT] Configuring Servos on Headers 006, 008, 017, 020...")
	legArr := must(tgservo.NewArray(pwmLeg))
	footArr := must(tgservo.NewArray(pwmFoot))

	llServo := servo.New180(must(legArr.Add(lLegPin)), 450, 2550)
	rlServo := servo.New180(must(legArr.Add(rLegPin)), 450, 2550)
	lfServo := servo.New360(must(footArr.Add(lFootPin)), 450, 2550)
	rfServo := servo.New360(must(footArr.Add(rFootPin)), 450, 2550)
	println("  Servos initialized.")

	// 3. INITIALIZE NINJA ROBOT
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
		println(">>> STARTING LOCOMOTION CYCLE #", cycle, "<<<")
		println("-------------------------------------------------------")

		// Phase 1: Buzzer Chime
		println("  [SOUND] Playing chime...")
		_ = bz.Tone(buzzer.Note{Period: buzzer.G4, Duration: 150 * time.Millisecond})
		_ = bz.Tone(buzzer.Note{Period: buzzer.C5, Duration: 200 * time.Millisecond})
		time.Sleep(300 * time.Millisecond)

		// Phase 2: Individual Servo Audit (identifies any unpowered servo)
		println("  [SERVOS] Individual motor verification:")
		println("    -> 1 beep: Motor 006 (Right Foot spin)...")
		beep(bz, 1)
		_ = rfServo.SetSpeed(35)
		time.Sleep(500 * time.Millisecond)
		_ = rfServo.SetSpeed(0)
		time.Sleep(200 * time.Millisecond)

		println("    -> 2 beeps: Motor 008 (Right Leg tilt)...")
		beep(bz, 2)
		_ = rlServo.SetAngle(65)
		time.Sleep(350 * time.Millisecond)
		_ = rlServo.SetAngle(115)
		time.Sleep(350 * time.Millisecond)
		_ = rlServo.SetAngle(90)
		time.Sleep(200 * time.Millisecond)

		println("    -> 3 beeps: Motor 017 (Left Leg tilt)...")
		beep(bz, 3)
		_ = llServo.SetAngle(65)
		time.Sleep(350 * time.Millisecond)
		_ = llServo.SetAngle(115)
		time.Sleep(350 * time.Millisecond)
		_ = llServo.SetAngle(90)
		time.Sleep(200 * time.Millisecond)

		println("    -> 4 beeps: Motor 020 (Left Foot spin)...")
		beep(bz, 4)
		_ = lfServo.SetSpeed(35)
		time.Sleep(500 * time.Millisecond)
		_ = lfServo.SetSpeed(0)
		time.Sleep(300 * time.Millisecond)

		// Phase 3: Coordinated Biped Walking
		println("\n  [WALK] Centering to Home position...")
		_ = n.Home()
		time.Sleep(500 * time.Millisecond)

		println("  [WALK] Walking 2 steps forward...")
		_ = n.Walk(2)
		time.Sleep(500 * time.Millisecond)

		println("  [WALK] Happy Wiggle celebration gesture...")
		_ = n.HappyWiggle()
		time.Sleep(500 * time.Millisecond)

		println("  [WALK] Moonwalking 2 steps backward...")
		_ = n.Moonwalk(2)
		time.Sleep(500 * time.Millisecond)

		println("\n[CYCLE #", cycle, " FINISHED] Resting 6 seconds...")
		time.Sleep(6 * time.Second)
		cycle++
	}
}
