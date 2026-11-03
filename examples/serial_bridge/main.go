// Package main implements the high-speed UART serial command bridge between
// the NiceNano (nRF52840) locomotion base and the ESP32-S3 robot brain.
//
// Wiring on red shield:
//   - Header 022 (machine.P0_22) -> ESP32-S3 TX (D4 / IO5)
//   - Header 024 (machine.P0_24) -> ESP32-S3 RX (D5 / IO6)
//   - GND                        -> ESP32-S3 GND
//   - Baud Rate: 115200, 8N1
package main

import (
	"bytes"
	"machine"
	"strconv"
	"strings"
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

	rLegPin  = board.RLegPin
	lLegPin  = board.LLegPin
	rFootPin = board.RFootPin
	lFootPin = board.LFootPin

	buzzerPin = board.BuzzerPin

	// Inter-board link UART (nRF52840 hardware UART0)
	uartLink = machine.UART0
	txPin    = machine.P0_24 // Header 024
	rxPin    = machine.P0_22 // Header 022
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func sendReply(uart *machine.UART, msg string) {
	println("[BRIDGE-OUT]:", msg)
	uart.Write([]byte(msg + "\n"))
}

func main() {
	time.Sleep(1 * time.Second)
	println("=== GOtto Ninja UART Bridge Initializing ===")

	// 1. Configure inter-board UART
	uartLink.Configure(machine.UARTConfig{
		TX:       txPin,
		RX:       rxPin,
		BaudRate: 115200,
	})

	// 2. Configure Buzzer
	bz := buzzer.New(buzzer.NewPwmChannel(pwmBuzzer, buzzerPin))
	if err := bz.Configure(); err != nil {
		println("Error configuring buzzer:", err.Error())
	}

	// 3. Power on switched VCC rail for servos & sensors
	board.EnableVcc()
	time.Sleep(150 * time.Millisecond)

	// 4. Configure Servos
	legArr := must(tgservo.NewArray(pwmLeg))
	footArr := must(tgservo.NewArray(pwmFoot))

	llServo := servo.New180(must(legArr.Add(lLegPin)), 450, 2550)
	rlServo := servo.New180(must(legArr.Add(rLegPin)), 450, 2550)
	lfServo := servo.New360(must(footArr.Add(lFootPin)), 450, 2550)
	rfServo := servo.New360(must(footArr.Add(rFootPin)), 450, 2550)

	n := ninja.New(rlServo, llServo, rfServo, lfServo, bz)

	// Apply Milan's calibrated trim values
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
	_ = n.Home()

	// 5. Configure HC-SR04 Ultrasonic Sensor
	us := hcsr04.New(board.USTrigPin, board.USEchoPin)
	us.Configure()

	// Signal ready to host / ESP32-S3
	sendReply(uartLink, "READY")

	// 6. Background Ultrasonic Proximity Monitor
	go func() {
		lastAlert := time.Now()
		for {
			distMm := us.ReadDistance()
			// If target detected within 350mm (and throttled to once every 2s)
			if distMm > 20 && distMm < 350 && time.Since(lastAlert) > 2*time.Second {
				lastAlert = time.Now()
				sendReply(uartLink, "PROXIMITY "+strconv.Itoa(int(distMm)))
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()

	// 7. Command Processing Loop
	var rxBuf bytes.Buffer
	for {
		for uartLink.Buffered() > 0 {
			b, err := uartLink.ReadByte()
			if err != nil {
				break
			}
			if b == '\n' || b == '\r' {
				if rxBuf.Len() > 0 {
					cmdLine := strings.TrimSpace(rxBuf.String())
					rxBuf.Reset()
					if len(cmdLine) > 0 {
						handleCommand(n, uartLink, cmdLine)
					}
				}
			} else {
				rxBuf.WriteByte(b)
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func handleCommand(n *ninja.Ninja, uart *machine.UART, line string) {
	println("[BRIDGE-IN]:", line)
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToUpper(parts[0])

	switch cmd {
	case "PING":
		sendReply(uart, "PONG")

	case "WAKEUP":
		_ = n.WakeUp()
		sendReply(uart, "DONE WAKEUP")

	case "SLEEP":
		_ = n.YawnAndSleep()
		sendReply(uart, "DONE SLEEP")

	case "HOME", "STOP":
		_ = n.Home()
		sendReply(uart, "DONE STOP")

	case "WIGGLE":
		_ = n.HappyWiggle()
		sendReply(uart, "DONE WIGGLE")

	case "CURIOUS":
		_ = n.CuriousLook()
		sendReply(uart, "DONE CURIOUS")

	case "WALK":
		steps := 2
		if len(parts) > 1 {
			if s, err := strconv.Atoi(parts[1]); err == nil && s > 0 {
				steps = s
			}
		}
		_ = n.Walk(steps)
		sendReply(uart, "DONE WALK")

	case "MOONWALK":
		steps := 2
		if len(parts) > 1 {
			if s, err := strconv.Atoi(parts[1]); err == nil && s > 0 {
				steps = s
			}
		}
		_ = n.Moonwalk(steps)
		sendReply(uart, "DONE MOONWALK")

	case "PIVOT":
		turns := 1
		if len(parts) > 1 {
			if t, err := strconv.Atoi(parts[1]); err == nil {
				turns = t
			}
		}
		_ = n.PivotTurn(turns)
		sendReply(uart, "DONE PIVOT")

	default:
		sendReply(uart, "ERR UNKNOWN_COMMAND")
	}
}
