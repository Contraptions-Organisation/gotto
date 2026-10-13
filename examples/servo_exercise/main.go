package main

import (
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
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func countBeeps(bz *buzzer.Buzzer, count int) {
	for i := 0; i < count; i++ {
		_ = bz.Tone(buzzer.Note{Period: buzzer.C5, Duration: 120 * time.Millisecond})
		time.Sleep(100 * time.Millisecond)
	}
	time.Sleep(300 * time.Millisecond)
}

func main() {
	time.Sleep(2 * time.Second)

	println("\n=======================================================")
	println("  GOTTO FULL-RANGE SERVO MOTION & BIPED LOCOMOTION     ")
	println("=======================================================")

	bz := buzzer.New(buzzer.NewPwmChannel(pwmBuzzer, buzzerPin))
	_ = bz.Configure()

	legArr := must(tgservo.NewArray(pwmLeg))
	footArr := must(tgservo.NewArray(pwmFoot))

	llServo := servo.New180(must(legArr.Add(lLegPin)), 450, 2550)
	rlServo := servo.New180(must(legArr.Add(rLegPin)), 450, 2550)
	lfServo := servo.New360(must(footArr.Add(lFootPin)), 450, 2550)
	rfServo := servo.New360(must(footArr.Add(rFootPin)), 450, 2550)

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
		println("\n>>> [CYCLE", cycle, "] RUNNING SERVO EXERCISE <<<")

		// ----------------------------------------------------
		// 1. MOTOR 006: RIGHT FOOT (360 continuous)
		// ----------------------------------------------------
		println("  -> [MOTOR 006] Right Foot: 1 Beep -> Fast Spin Forward & Reverse...")
		countBeeps(bz, 1)
		_ = rfServo.SetSpeed(60)
		time.Sleep(1200 * time.Millisecond)
		_ = rfServo.SetSpeed(-60)
		time.Sleep(1200 * time.Millisecond)
		_ = rfServo.SetSpeed(0)
		time.Sleep(500 * time.Millisecond)

		// ----------------------------------------------------
		// 2. MOTOR 008: RIGHT LEG (180 degree hip)
		// ----------------------------------------------------
		println("  -> [MOTOR 008] Right Leg: 2 Beeps -> Large Sweep [90 -> 45 -> 135 -> 90]...")
		countBeeps(bz, 2)
		for _, a := range []int{90, 45, 135, 45, 135, 90} {
			_ = rlServo.SetAngle(a)
			time.Sleep(450 * time.Millisecond)
		}
		time.Sleep(500 * time.Millisecond)

		// ----------------------------------------------------
		// 3. MOTOR 017: LEFT LEG (180 degree hip)
		// ----------------------------------------------------
		println("  -> [MOTOR 017] Left Leg: 3 Beeps -> Large Sweep [90 -> 45 -> 135 -> 90]...")
		countBeeps(bz, 3)
		for _, a := range []int{90, 45, 135, 45, 135, 90} {
			_ = llServo.SetAngle(a)
			time.Sleep(450 * time.Millisecond)
		}
		time.Sleep(500 * time.Millisecond)

		// ----------------------------------------------------
		// 4. MOTOR 020: LEFT FOOT (360 continuous)
		// ----------------------------------------------------
		println("  -> [MOTOR 020] Left Foot: 4 Beeps -> Fast Spin Forward & Reverse...")
		countBeeps(bz, 4)
		_ = lfServo.SetSpeed(60)
		time.Sleep(1200 * time.Millisecond)
		_ = lfServo.SetSpeed(-60)
		time.Sleep(1200 * time.Millisecond)
		_ = lfServo.SetSpeed(0)
		time.Sleep(500 * time.Millisecond)

		// ----------------------------------------------------
		// 5. COORDINATED BIPED WALKING
		// ----------------------------------------------------
		println("\n  -> [BIPED GAIT] Home Posture & Walking 3 Steps Forward...")
		_ = bz.Tone(buzzer.Note{Period: buzzer.G4, Duration: 150 * time.Millisecond})
		_ = bz.Tone(buzzer.Note{Period: buzzer.C5, Duration: 250 * time.Millisecond})
		_ = n.Home()
		time.Sleep(500 * time.Millisecond)

		_ = n.Walk(3)
		time.Sleep(500 * time.Millisecond)

		println("  -> [BIPED GAIT] Happy Wiggle Celebration...")
		_ = n.HappyWiggle()
		time.Sleep(500 * time.Millisecond)

		println("  -> [BIPED GAIT] Walking 3 Steps Backward (Moonwalk)...")
		_ = n.Moonwalk(3)
		time.Sleep(500 * time.Millisecond)

		println("\n[CYCLE", cycle, "COMPLETE] Resting 5 seconds before next cycle...")
		time.Sleep(5 * time.Second)
		cycle++
	}
}
