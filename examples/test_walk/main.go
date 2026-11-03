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

	rLegPin  = board.RLegPin
	lLegPin  = board.LLegPin
	rFootPin = board.RFootPin
	lFootPin = board.LFootPin

	buzzerPin = board.BuzzerPin
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func main() {
	time.Sleep(2 * time.Second)
	println("GOtto test: Buzzer and Walk starting...")

	// 1. Initialize Buzzer on board.BuzzerPin (P1.13) using PWM0
	bz := buzzer.New(buzzer.NewPwmChannel(pwmBuzzer, buzzerPin))
	if err := bz.Configure(); err != nil {
		println("Error configuring buzzer:", err.Error())
		return
	}

	println("Testing Buzzer: Playing startup chime...")
	startupMelody := []buzzer.Note{
		{Period: buzzer.C4, Duration: 150 * time.Millisecond},
		{Period: buzzer.E4, Duration: 150 * time.Millisecond},
		{Period: buzzer.G4, Duration: 150 * time.Millisecond},
		{Period: buzzer.C5, Duration: 300 * time.Millisecond},
	}
	bz.PlayMelody(startupMelody)

	time.Sleep(1 * time.Second)

	// 2. Enable VCC power rail for servos
	println("Enabling servo VCC rail...")
	board.EnableVcc()
	time.Sleep(300 * time.Millisecond)

	// 3. Configure Leg (180 deg) and Foot (360 deg) servos
	println("Configuring servos...")
	legArr := must(tgservo.NewArray(pwmLeg))
	footArr := must(tgservo.NewArray(pwmFoot))

	llServo := servo.New180(must(legArr.Add(lLegPin)), 450, 2550)
	rlServo := servo.New180(must(legArr.Add(rLegPin)), 450, 2550)
	lfServo := servo.New360(must(footArr.Add(lFootPin)), 450, 2550)
	rfServo := servo.New360(must(footArr.Add(rFootPin)), 450, 2550)

	// 4. Initialize Ninja robot
	n := ninja.New(rlServo, llServo, rfServo, lfServo, bz)

	// Standard trim values from repo examples
	n.Trim(ninja.Trim{
		TiltAngle:         0,
		LeftStepDuration:  0,
		RightStepDuration: 150 * time.Millisecond,
		LfSpeed:           0,
		RfSpeed:           0,
		LlAngle:           20,
		RlAngle:           12,
	})

	// 5. Stand / Home position
	println("Setting home position (standing straight)...")
	if err := n.Mode(ninja.ModeWalk); err != nil {
		println("Error setting walk mode:", err.Error())
		return
	}
	time.Sleep(1 * time.Second)

	// Beep before walking
	_ = bz.Tone(buzzer.Note{Period: buzzer.A4, Duration: 200 * time.Millisecond})
	time.Sleep(500 * time.Millisecond)

	// 6. Test Walking (2 steps forward)
	println("Walking 2 steps forward...")
	if err := n.Walk(2); err != nil {
		println("Error walking:", err.Error())
	}

	time.Sleep(1 * time.Second)

	// 7. Success tone
	println("Walk test finished! Playing completion tone...")
	doneMelody := []buzzer.Note{
		{Period: buzzer.G4, Duration: 150 * time.Millisecond},
		{Period: buzzer.C5, Duration: 350 * time.Millisecond},
	}
	bz.PlayMelody(doneMelody)

	for {
		time.Sleep(time.Hour)
	}
}
