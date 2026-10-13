package main

import (
	"machine"
	"time"

	"github.com/HattoriHanzo031/gotto/board"
	"github.com/HattoriHanzo031/gotto/buzzer"
	"github.com/HattoriHanzo031/gotto/ninja"
	"github.com/HattoriHanzo031/gotto/remote"
	"github.com/HattoriHanzo031/gotto/servo"

	"tinygo.org/x/drivers/hcsr04"
	tgservo "tinygo.org/x/drivers/servo"
)

var (
	footPwm   = board.FootPWM
	legPwm    = board.LegPWM
	buzzerPwm = board.BuzzerPWM

	rLeg  = board.RLegPin
	lLeg  = board.LLegPin
	rFoot = board.RFootPin
	lFoot = board.LFootPin

	usTrig = board.USTrigPin
	usEcho = board.USEchoPin

	buzzerPin = board.BuzzerPin

	buttonPin = board.ButtonPin
)

func main() {
	time.Sleep(3 * time.Second)

	board.EnableVcc()

	legArr := must(tgservo.NewArray(legPwm))
	footArr := must(tgservo.NewArray(footPwm))

	llServo := servo.New180(must(legArr.Add(lLeg)), 450, 2550)
	rlServo := servo.New180(must(legArr.Add(rLeg)), 450, 2550)
	lfServo := servo.New360(must(footArr.Add(lFoot)), 450, 2550)
	rfServo := servo.New360(must(footArr.Add(rFoot)), 450, 2550)

	bz := buzzer.New(buzzer.NewPwmChannel(buzzerPwm, buzzerPin))
	err := bz.Configure()
	if err != nil {
		panic(err)
	}

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

	// Initialize ultrasonic sensor
	us := hcsr04.New(usTrig, usEcho)
	us.Configure()

	// Set custom commands for obstacle avoidance
	_ = n.SetCustomCommand(0, obstacleAvoidanceWalkFn(us))
	_ = n.SetCustomCommand(1, obstacleAvoidanceRollFn(us))

	time.Sleep(500 * time.Millisecond)

	// Play a tone to indicate the robot is ready
	n.BuzzerTone(buzzer.Note{
		Period:   buzzer.B4,
		Duration: time.Duration(time.Second),
	})

	rmt := NewMicroBlue()
	if err := rmt.Start(); err != nil {
		println("Error starting remote:", err.Error())
		return
	}

	commandCh := make(chan remote.Command, 1)

	buttonCommand := remote.Command{
		Op:   remote.OpSetMode,
		Args: [2]int{0, 0},
	}

	t := time.Now()
	buttonPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	buttonPin.SetInterrupt(machine.PinToggle, func(p machine.Pin) {
		if time.Now().Before(t.Add(time.Second)) {
			return
		}
		t = time.Now()

		select {
		case commandCh <- buttonCommand:
			if buttonPin.Get() {
				buttonCommand.Args[0] = 1
			} else {
				buttonCommand.Args[0] = 0
			}
		default:
		}
	})

	go func() {
		for {
			// Non-blocking send to skip the command if previous command is not yet processed
			select {
			case commandCh <- rmt.ReadCommand():
			default:
			}
		}
	}()

	for cmd := range commandCh {
		if err := cmd.Execute(n); err != nil {
			println("Error executing command:", err.Error())
		}
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
