package main

import (
	"bytes"
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

	rLeg  = board.RLegPin
	lLeg  = board.LLegPin
	rFoot = board.RFootPin
	lFoot = board.LFootPin

	buzzerPin = board.BuzzerPin
)

func main() {
	machine.InitSerial()
	time.Sleep(3 * time.Second)

	board.EnableVcc()

	legArr := must(tgservo.NewArray(pwmLeg))
	footArr := must(tgservo.NewArray(pwmFoot))

	llServo := servo.New180(must(legArr.Add(lLeg)), 450, 2550)
	rlServo := servo.New180(must(legArr.Add(rLeg)), 450, 2550)
	lfServo := servo.New360(must(footArr.Add(lFoot)), 450, 2550)
	rfServo := servo.New360(must(footArr.Add(rFoot)), 450, 2550)

	bz := buzzer.New(buzzer.NewPwmChannel(pwmBuzzer, buzzerPin))
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
	n.Mode(ninja.ModeWalk)
	n.Home()

	var commandBuffer [255]byte

	for {
		command := readCommand(commandBuffer[:0])
		println("Got", string(command))
		parts := bytes.Fields(command)

		if len(parts) == 0 {
			continue
		}

		switch string(parts[0]) {
		case "ll+":
			trim.LlAngle++
			println("left leg angle trim:", trim.LlAngle)
			n.Trim(trim)
			testLegs(n)
		case "ll-":
			if trim.LlAngle == 0 {
				println("only positive values allowed")
				continue
			}
			trim.LlAngle--
			println("left leg angle trim:", trim.LlAngle)
			n.Trim(trim)
			testLegs(n)
		case "rl+":
			trim.RlAngle++
			println("right leg angle trim:", trim.RlAngle)
			n.Trim(trim)
			testLegs(n)
		case "rl-":
			if trim.RlAngle == 0 {
				println("only positive values allowed")
				continue
			}
			trim.RlAngle--
			println("right leg angle trim:", trim.RlAngle)
			n.Trim(trim)
			testLegs(n)
		case "lf+":
			trim.LfSpeed++
			println("left foot speed trim:", trim.LfSpeed)
			n.Trim(trim)
			testRolling(n)
		case "lf-":
			trim.LfSpeed--
			println("left foot speed trim:", trim.LfSpeed)
			n.Trim(trim)
			testRolling(n)
		case "rf+":
			trim.RfSpeed++
			println("right foot speed trim:", trim.RfSpeed)
			n.Trim(trim)
			testRolling(n)
		case "rf-":
			trim.RfSpeed--
			println("right foot speed trim:", trim.RfSpeed)
			n.Trim(trim)
			testRolling(n)
		case "tilt+":
			trim.TiltAngle++
			println("tilt angle trim:", trim.TiltAngle)
			n.Trim(trim)
			testTilting(n)
		case "tilt-":
			trim.TiltAngle--
			println("tilt angle trim:", trim.TiltAngle)
			n.Trim(trim)
			testTilting(n)
		case "reset":
			trim = ninja.Trim{}
			n.Trim(trim)
			println("reset trims")
			n.Home()
		case "roll":
			err := n.Mode(ninja.ModeRoll)
			if err != nil {
				println(err)
			}
		case "walk":
			err := n.Mode(ninja.ModeWalk)
			if err != nil {
				println(err)
			}
		case "demo":
			fullTest(n)
			println("Demo complete")
		default:
			println("unknown command")
		}
	}

}

func readCommand(buffer []byte) []byte {
	buffer = buffer[:0]
	for {
		// Check if any data is available to read from the serial port
		if machine.Serial.Buffered() == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		// Read a single byte
		data, err := machine.Serial.ReadByte()
		if err != nil {
			println("Error reading from serial:", err)
			continue
		}

		// Process the received byte
		if data == '\r' || data == '\n' {
			// Handle newline or carriage return (e.g., end of input)
			return buffer
		}
		// Echo the character back to the serial monitor
		machine.Serial.WriteByte(data)
		buffer = append(buffer, data)
	}
}

func testLegs(n *ninja.Ninja) {
	n.Mode(ninja.ModeWalk)
	n.Home()
}

func testWalking(n *ninja.Ninja) {
	println("Walking test")
	n.Mode(ninja.ModeWalk)
	n.Home()
	time.Sleep(300 * time.Millisecond)

	n.Walk(3)
	time.Sleep(2 * time.Second)
	n.Walk(-3)
	time.Sleep(2 * time.Second)

	println("Walking test complete")
}

func testRolling(n *ninja.Ninja) {
	println("Rolling test")
	n.Mode(ninja.ModeRoll)
	n.Home()
	time.Sleep(300 * time.Millisecond)

	n.Roll(50, 0)
	time.Sleep(2 * time.Second)
	n.RollStop()
	time.Sleep(2 * time.Second)
	n.Roll(-50, 0)
	time.Sleep(2 * time.Second)
	n.RollStop()
	time.Sleep(2 * time.Second)
	n.Roll(0, 50)
	time.Sleep(2 * time.Second)
	n.RollStop()

	println("Rolling test complete")
}

func testTilting(n *ninja.Ninja) {
	println("Tilting test")
	n.Mode(ninja.ModeWalk)
	n.Home()
	time.Sleep(300 * time.Millisecond)

	println("Tilt Left")
	n.Tilt(ninja.TiltLeft)
	time.Sleep(2 * time.Second)

	println("Tilt Center")
	n.Tilt(ninja.TiltReturnFromLeft)
	time.Sleep(2 * time.Second)

	println("Tilt Right")
	n.Tilt(ninja.TiltRight)
	time.Sleep(2 * time.Second)

	println("Tilt Center")
	n.Tilt(ninja.TiltReturnFromRight)

	println("Tilting test complete")
}

func fullTest(n *ninja.Ninja) {
	testTilting(n)
	time.Sleep(2 * time.Second)
	testWalking(n)
	time.Sleep(2 * time.Second)
	testRolling(n)
	time.Sleep(2 * time.Second)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
