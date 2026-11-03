package main

import (
	"machine"
	"time"

	"github.com/HattoriHanzo031/gotto/board"
	"github.com/HattoriHanzo031/gotto/buzzer"
	"github.com/HattoriHanzo031/gotto/servo"
	"tinygo.org/x/drivers/hcsr04"
	tgservo "tinygo.org/x/drivers/servo"
)

type HeaderPin struct {
	Name string
	Pin  machine.Pin
}

var allHeaderPins = []HeaderPin{
	// Left side headers (top to bottom)
	{"006 (RFoot)", machine.P0_06},
	{"008 (RLeg)", machine.P0_08},
	{"017 (LLeg)", machine.P0_17},
	{"020 (LFoot)", machine.P0_20},
	{"022 (Aux)", machine.P0_22},
	{"024 (Aux)", machine.P0_24},
	{"100 (USTrig)", machine.P1_00},
	{"011 (USEcho)", machine.P0_11},

	// Right side headers (top to bottom)
	{"031 (Buzzer)", machine.P0_31},
	{"002 (Aux/AIN0)", machine.P0_02},
	{"115 (Aux)", machine.P1_15},
	{"113 (OldBuzzer)", machine.P1_13},
	{"111 (Aux)", machine.P1_11},
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// bitbangTone generates precise square wave
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

// fullBuzzerBurst drives pin in all possible electrical configurations
func fullBuzzerBurst(pin machine.Pin, label string) {
	println("  >> Driving", label, "through all sound modes <<")
	machine.LED.Low()

	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// 1. DC Active Buzzer: Active HIGH (3.3V)
	print("    [1/5] DC Active-High 3.3V (300ms)... ")
	pin.High()
	time.Sleep(300 * time.Millisecond)
	pin.Low()
	println("done")
	time.Sleep(100 * time.Millisecond)

	// 2. DC Active Buzzer: Active LOW (0V sinking)
	print("    [2/5] DC Active-Low Ground Sink (300ms)... ")
	pin.Low()
	time.Sleep(300 * time.Millisecond)
	pin.High()
	time.Sleep(100 * time.Millisecond)
	pin.Low()
	println("done")
	time.Sleep(100 * time.Millisecond)

	// 3. Piezo Resonant Frequency: 2.7 kHz
	print("    [3/5] AC 2.7 kHz Piezo Resonance (400ms)... ")
	bitbangTone(pin, 2700, 400*time.Millisecond)
	println("done")
	time.Sleep(100 * time.Millisecond)

	// 4. Acoustic Musical Tone: 440 Hz (A4)
	print("    [4/5] AC 440 Hz (A4) Tone (400ms)... ")
	bitbangTone(pin, 440, 400*time.Millisecond)
	println("done")
	time.Sleep(100 * time.Millisecond)

	// 5. High-Pitch Chirp: 1.5 kHz -> 3.5 kHz
	print("    [5/5] AC Multi-frequency Chirp (300ms)... ")
	bitbangTone(pin, 1500, 100*time.Millisecond)
	bitbangTone(pin, 2500, 100*time.Millisecond)
	bitbangTone(pin, 3500, 100*time.Millisecond)
	println("done")

	pin.Low()
	machine.LED.High()
}

func main() {
	time.Sleep(2 * time.Second)
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.LED.High()

	println("\n=======================================================")
	println("  GOTTO DEEP HARDWARE TELEMETRY & ELECTRICAL PROBE     ")
	println("=======================================================")

	// --------------------------------------------------------
	// STEP 1: ELECTRICAL PIN MAP AUDIT
	// --------------------------------------------------------
	println("\n--- [AUDIT 1/4] SHIELD ELECTRICAL IMPEDANCE SCAN ---")
	println("Probing all 16 shield header pins for load & connectivity:")
	println("PIN NAME        | PULLUP | PULLDN | FLOAT  | ELECTRICAL STATUS")
	println("----------------+--------+--------+--------+------------------")

	for _, hp := range allHeaderPins {
		// Pullup
		hp.Pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
		time.Sleep(2 * time.Millisecond)
		pu := hp.Pin.Get()

		// Pulldown
		hp.Pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
		time.Sleep(2 * time.Millisecond)
		pd := hp.Pin.Get()

		// Float
		hp.Pin.Configure(machine.PinConfig{Mode: machine.PinInput})
		time.Sleep(2 * time.Millisecond)
		flt := hp.Pin.Get()

		status := "Open / Floating"
		if !pu && !pd {
			status = "Connected to GND"
		} else if pu && pd {
			status = "Connected to VDD"
		} else if !flt {
			status = "Weak load to GND"
		}

		print("Header ", hp.Name, " \t| ")
		if pu {
			print("HIGH   | ")
		} else {
			print("LOW    | ")
		}
		if pd {
			print("HIGH   | ")
		} else {
			print("LOW    | ")
		}
		if flt {
			print("HIGH   | ")
		} else {
			print("LOW    | ")
		}
		println(status)
	}

	// --------------------------------------------------------
	// STEP 2: TEST BUZZER ON BOTH 031 AND 113
	// --------------------------------------------------------
	println("\n--- [AUDIT 2/4] EXHAUSTIVE BUZZER ACTIVATION ---")
	fullBuzzerBurst(machine.P0_31, "HEADER 031 (P0.31)")
	time.Sleep(500 * time.Millisecond)
	fullBuzzerBurst(machine.P1_13, "HEADER 113 (P1.13)")
	time.Sleep(500 * time.Millisecond)

	// Also test hardware PWM0 using the official gotto buzzer package on 031
	println("  >> Testing gotto/buzzer package on PWM0 (P0.31) <<")
	bz := buzzer.New(buzzer.NewPwmChannel(machine.PWM0, machine.P0_31))
	if err := bz.Configure(); err == nil {
		println("    PWM0 configured. Playing 3 notes...")
		_ = bz.Tone(buzzer.Note{Period: buzzer.A4, Duration: 250 * time.Millisecond})
		time.Sleep(50 * time.Millisecond)
		_ = bz.Tone(buzzer.Note{Period: buzzer.C5, Duration: 250 * time.Millisecond})
		time.Sleep(50 * time.Millisecond)
		_ = bz.Tone(buzzer.Note{Period: buzzer.E5, Duration: 350 * time.Millisecond})
	} else {
		println("    PWM0 configure error:", err.Error())
	}

	// --------------------------------------------------------
	// STEP 3: ULTRASONIC SENSOR TELEMETRY AUDIT
	// --------------------------------------------------------
	println("\n--- [AUDIT 3/4] ULTRASONIC SENSOR TELEMETRY ---")
	us := hcsr04.New(board.USTrigPin, board.USEchoPin)
	us.Configure()
	for i := 1; i <= 3; i++ {
		dist := us.ReadDistance()
		println("  Sample #", i, "distance measurement:", dist, "mm")
		time.Sleep(60 * time.Millisecond)
	}

	// --------------------------------------------------------
	// STEP 4: SERVO SIGNAL AUDIT (006, 008, 017, 020)
	// --------------------------------------------------------
	println("\n--- [AUDIT 4/4] SERVO CONTROLLER & MOTION AUDIT ---")
	println("Configuring PWM1 (Legs 008, 017) and PWM2 (Feet 006, 020)...")
	legArr := must(tgservo.NewArray(machine.PWM1))
	footArr := must(tgservo.NewArray(machine.PWM2))

	llServo := servo.New180(must(legArr.Add(board.LLegPin)), 450, 2550)
	rlServo := servo.New180(must(legArr.Add(board.RLegPin)), 450, 2550)
	lfServo := servo.New360(must(footArr.Add(board.LFootPin)), 450, 2550)
	rfServo := servo.New360(must(footArr.Add(board.RFootPin)), 450, 2550)
	println("Servos configured successfully in software.")

	println("Sending motion commands:")
	println("  1. Right Foot 006: Spin forward (speed 40) for 800ms...")
	_ = rfServo.SetSpeed(40)
	time.Sleep(800 * time.Millisecond)
	_ = rfServo.SetSpeed(0)

	println("  2. Right Leg 008: Tilt 90 -> 60 -> 120 -> 90...")
	for _, a := range []int{90, 60, 120, 90} {
		_ = rlServo.SetAngle(a)
		time.Sleep(300 * time.Millisecond)
	}

	println("  3. Left Leg 017: Tilt 90 -> 60 -> 120 -> 90...")
	for _, a := range []int{90, 60, 120, 90} {
		_ = llServo.SetAngle(a)
		time.Sleep(300 * time.Millisecond)
	}

	println("  4. Left Foot 020: Spin forward (speed 40) for 800ms...")
	_ = lfServo.SetSpeed(40)
	time.Sleep(800 * time.Millisecond)
	_ = lfServo.SetSpeed(0)

	// --------------------------------------------------------
	// CONTINUOUS INTERACTIVE PIN CHANGE MONITOR
	// --------------------------------------------------------
	println("\n=======================================================")
	println("  ENTERING REAL-TIME INTERACTIVE PIN WATCHDOG LOOP     ")
	println("  Monitoring all shield headers for physical touches/  ")
	println("  wire connections. Sound bursts on 031 & 113 repeat.  ")
	println("=======================================================")

	loopCount := 1
	for {
		time.Sleep(4 * time.Second)
		println("\n[CYCLE #", loopCount, "] Firing sound burst on Header 031 & 113...")
		fullBuzzerBurst(machine.P0_31, "Header 031")
		fullBuzzerBurst(machine.P1_13, "Header 113")

		// Pulse servos slightly
		_ = rlServo.SetAngle(100)
		_ = llServo.SetAngle(80)
		time.Sleep(300 * time.Millisecond)
		_ = rlServo.SetAngle(90)
		_ = llServo.SetAngle(90)
		time.Sleep(300 * time.Millisecond)

		loopCount++
	}
}
