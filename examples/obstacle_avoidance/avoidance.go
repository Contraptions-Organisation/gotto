package main

import (
	"math/rand/v2"
	"time"

	"github.com/HattoriHanzo031/gotto/board"
	"github.com/HattoriHanzo031/gotto/buzzer"
	"github.com/HattoriHanzo031/gotto/ninja"
	"github.com/HattoriHanzo031/gotto/servo"
	"tinygo.org/x/drivers/hcsr04"
	tgservo "tinygo.org/x/drivers/servo"
)

var (
	footPwm   = board.FootPWM
	legPwm    = board.LegPWM
	buzzerPwm = board.BuzzerPWM

	rLegPin  = board.RLegPin
	lLegPin  = board.LLegPin
	rFootPin = board.RFootPin
	lFootPin = board.LFootPin

	usTrigPin = board.USTrigPin
	usEchoPin = board.USEchoPin

	buzzerPin = board.BuzzerPin
)

func main() {
	time.Sleep(3 * time.Second)

	board.EnableVcc()

	legArr := must(tgservo.NewArray(legPwm))
	footArr := must(tgservo.NewArray(footPwm))

	llServo := servo.New180(must(legArr.Add(lLegPin)), 450, 2550)
	rlServo := servo.New180(must(legArr.Add(rLegPin)), 450, 2550)
	lfServo := servo.New360(must(footArr.Add(lFootPin)), 450, 2550)
	rfServo := servo.New360(must(footArr.Add(rFootPin)), 450, 2550)

	bz := buzzer.New(buzzer.NewPwmChannel(buzzerPwm, buzzerPin))
	err := bz.Configure()
	if err != nil {
		panic(err)
	}

	n := ninja.New(rlServo, llServo, rfServo, lfServo, bz)
	trim := ninja.Trim{
		TiltAngle:         0,
		LeftStepDuration:  0,
		RightStepDuration: 150 * time.Millisecond,
		LfSpeed:           0,
		RfSpeed:           0,
		LlAngle:           20,
		RlAngle:           12,
	}

	n.Trim(trim)

	us := hcsr04.New(usTrigPin, usEchoPin)
	us.Configure()

	for {
		// Walk 2 minutes with obstacle avoidance
		start := time.Now()
		n.Mode(ninja.ModeWalk)
		for time.Since(start) < 2*time.Minute {
			// If an obstacle is detected within 150mm, step back and spin random amount to avoid it
			dist := us.ReadDistance()
			if dist != 0 && dist < 150 {
				n.Walk(-1)
				n.RightLegSpin(20, 100*time.Duration(rand.IntN(9)+6)*time.Millisecond)
				continue
			}
			n.Walk(1)
		}

		// Roll 2 minutes with obstacle avoidance
		start = time.Now()
		n.Mode(ninja.ModeRoll)
		n.Roll(50, 0)
		for time.Since(start) < 2*time.Minute {
			// If an obstacle is detected within 150mm, step back and spin random amount to avoid it
			dist := us.ReadDistance()
			if dist != 0 && dist < 150 {
				n.RollStop()
				time.Sleep(500 * time.Millisecond)
				n.Roll(-50, 0)
				time.Sleep(500 * time.Millisecond)
				n.Roll(0, 50)
				time.Sleep(100 * time.Duration(rand.IntN(9)+1) * time.Millisecond)
				n.Roll(50, 0)
			}
		}
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
