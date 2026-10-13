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

	rLegPin  = board.RLegPin
	lLegPin  = board.LLegPin
	rFootPin = board.RFootPin
	lFootPin = board.LFootPin

	buzzerPin = board.BuzzerPin
	led       = machine.LED
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func main() {
	machine.InitSerial()
	time.Sleep(1 * time.Second)

	// Power on the switched VCC rail
	board.EnableVcc()
	time.Sleep(200 * time.Millisecond)

	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.Low() // LED on

	println("\n=======================================================")
	println("  OTTONINJA CONTINUOUS WALKING FIRMWARE                ")
	println("=======================================================")

	bz := buzzer.New(buzzer.NewPwmChannel(pwmBuzzer, buzzerPin))
	if err := bz.Configure(); err == nil {
		_ = bz.Tone(buzzer.Note{Period: buzzer.C5, Duration: 150 * time.Millisecond})
		_ = bz.Tone(buzzer.Note{Period: buzzer.E5, Duration: 250 * time.Millisecond})
	}

	println("Configuring servos...")
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

	println("Centering to Home position (both feet planted flat)...")
	if err := n.Mode(ninja.ModeWalk); err != nil {
		println("Error setting walk mode:", err.Error())
	}
	_ = n.Home()
	time.Sleep(1500 * time.Millisecond)

	println("Starting controlled forward walk (both feet planted between steps)...")
	step := 1
	for {
		led.Low()
		println(">>> Walking step", step)
		if err := n.Walk(1); err != nil {
			println("Walk error:", err.Error())
		}
		led.High()
		// Stable pause with both feet flat on the ground
		time.Sleep(400 * time.Millisecond)
		step++
	}
}
