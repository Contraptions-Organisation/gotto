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
	println("GOtto Autonomous Personality Engine initialized.")

	// Buzzer
	bz := buzzer.New(buzzer.NewPwmChannel(pwmBuzzer, buzzerPin))
	if err := bz.Configure(); err != nil {
		println("Error configuring buzzer:", err.Error())
		return
	}

	// Enable VCC servo rail
	board.EnableVcc()
	time.Sleep(200 * time.Millisecond)

	// Servos
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

	// Initial stand
	_ = n.Mode(ninja.ModeWalk)
	_ = n.Home()

	for {
		// 1. Wake up sequence
		println("State: Waking up...")
		_ = n.WakeUp()
		time.Sleep(1 * time.Second)

		// 2. Walk forward
		println("State: Walking forward...")
		_ = n.Walk(2)
		time.Sleep(800 * time.Millisecond)

		// 3. Curious observation
		println("State: Looking around curiously...")
		_ = n.CuriousLook()
		time.Sleep(800 * time.Millisecond)

		// 4. Pivot Turn
		println("State: Pivot turning right...")
		_ = n.PivotTurn(2)
		time.Sleep(800 * time.Millisecond)

		// 5. Happy dance
		println("State: Performing Happy Wiggle...")
		_ = n.HappyWiggle()
		time.Sleep(1 * time.Second)

		// 6. Moonwalk backward
		println("State: Moonwalking backward...")
		_ = n.Moonwalk(2)
		time.Sleep(800 * time.Millisecond)

		// 7. Pivot Turn back
		println("State: Pivot turning left...")
		_ = n.PivotTurn(-2)
		time.Sleep(1 * time.Second)

		// 8. Sleep / Power-saving rest
		println("State: Power-saving nap (cutting servo VCC)...")
		_ = n.YawnAndSleep()

		// Rest quietly for 10 seconds before next autonomous cycle
		time.Sleep(10 * time.Second)
	}
}
