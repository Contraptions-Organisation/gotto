package main

import (
	"time"

	"github.com/HattoriHanzo031/gotto/board"
	"github.com/HattoriHanzo031/gotto/buzzer"
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

// beepPitch is the single pitch used for every BIOS-style beep code below.
// Only the beep *count* carries meaning, like a motherboard POST code.
const beepPitch = buzzer.C5

func main() {
	time.Sleep(3 * time.Second)
	println("GOtto hardware diagnostic demo starting...")

	bz := buzzer.New(buzzer.NewPwmChannel(pwmBuzzer, buzzerPin))
	if err := bz.Configure(); err != nil {
		panic(err)
	}

	// 1 beep: board booted, buzzer works.
	println("Beep code 1: board booted")
	beep(bz, 1)
	time.Sleep(500 * time.Millisecond)

	println("Enabling servo power (VCC rail)")
	board.EnableVcc()
	time.Sleep(200 * time.Millisecond)

	// 2 beeps: VCC rail enabled, about to configure servos.
	println("Beep code 2: servo power on")
	beep(bz, 2)
	time.Sleep(500 * time.Millisecond)

	legArr := must(tgservo.NewArray(pwmLeg))
	footArr := must(tgservo.NewArray(pwmFoot))

	llServo := servo.New180(must(legArr.Add(lLegPin)), 450, 2550)
	rlServo := servo.New180(must(legArr.Add(rLegPin)), 450, 2550)
	lfServo := servo.New360(must(footArr.Add(lFootPin)), 450, 2550)
	rfServo := servo.New360(must(footArr.Add(rFootPin)), 450, 2550)

	// 3 beeps: Left Leg, then silence while it actually moves.
	println("Beep code 3: testing Left Leg")
	beep(bz, 3)
	testLegServo("Left Leg", llServo)

	// 4 beeps: Right Leg.
	println("Beep code 4: testing Right Leg")
	beep(bz, 4)
	testLegServo("Right Leg", rlServo)

	// 5 beeps: Left Foot.
	println("Beep code 5: testing Left Foot")
	beep(bz, 5)
	testFootServo("Left Foot", lfServo)

	// 6 beeps: Right Foot.
	println("Beep code 6: testing Right Foot")
	beep(bz, 6)
	testFootServo("Right Foot", rfServo)

	// Long steady tone: diagnostics finished.
	println("Diagnostics finished. Restart the board to rerun.")
	_ = bz.Tone(buzzer.Note{Period: beepPitch, Duration: 1 * time.Second})
	for {
		time.Sleep(time.Hour)
	}
}

// Helper to avoid boilerplate error handling in main
// Not recommended for production code
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// beep plays `count` short, evenly spaced beeps at a fixed pitch, BIOS POST
// style: the number of beeps identifies what's about to happen, and nothing
// physical moves until the beeps are done.
func beep(bz *buzzer.Buzzer, count int) {
	tone := buzzer.Note{Period: beepPitch, Duration: 150 * time.Millisecond}
	rest := buzzer.Note{Period: buzzer.Silence, Duration: 150 * time.Millisecond}
	for i := 0; i < count; i++ {
		_ = bz.Tone(tone)
		_ = bz.Tone(rest)
	}
}

func testLegServo(label string, s servo.Servo180) {
	println(label, "servo sweep")
	angles := [...]int{90, 60, 120, 90}
	for _, angle := range angles {
		if err := s.SetAngle(angle); err != nil {
			println(label, "error:", err.Error())
			return
		}
		time.Sleep(700 * time.Millisecond)
	}
}

func testFootServo(label string, s servo.Servo360) {
	println(label, "speed sweep")
	speeds := [...]int{40, -40, 0}
	for _, speed := range speeds {
		if err := s.SetSpeed(speed); err != nil {
			println(label, "error:", err.Error())
			return
		}
		time.Sleep(800 * time.Millisecond)
	}
	_ = s.SetSpeed(0)
}
