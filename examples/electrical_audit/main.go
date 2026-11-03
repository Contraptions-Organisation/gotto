package main

import (
	"machine"
	"time"

	"github.com/HattoriHanzo031/gotto/board"
)

type pinProbe struct {
	name string
	pin  machine.Pin
}

var allHeaderPins = []pinProbe{
	{"P1.13 (D113)", machine.P1_13},
	{"P0.31 (D031)", machine.P0_31},
	{"P0.29 (D029)", machine.P0_29},
	{"P0.02 (D002)", machine.P0_02},
	{"P1.15 (D115)", machine.P1_15},
	{"P1.11 (D111)", machine.P1_11},
	{"P0.22 (D022)", machine.P0_22},
	{"P0.24 (D024)", machine.P0_24},
	{"P1.00 (D100)", machine.P1_00},
	{"P0.11 (D011)", machine.P0_11},
	{"P1.04 (D104)", machine.P1_04},
	{"P1.06 (D106)", machine.P1_06},
	{"P1.01 (D101)", machine.P1_01},
	{"P1.02 (D102)", machine.P1_02},
	{"P1.07 (D107)", machine.P1_07},
	{"P0.10 (D010)", machine.P0_10},
	{"P0.09 (D009)", machine.P0_09},
	{"P0.17 (D017 - LLeg)", machine.P0_17},
	{"P0.08 (D008 - RLeg)", machine.P0_08},
	{"P0.20 (D020 - LFoot)", machine.P0_20},
	{"P0.06 (D006 - RFoot)", machine.P0_06},
}

func main() {
	time.Sleep(2 * time.Second)
	println("\n=======================================================")
	println("     ELECTRICAL CONTINUITY & IMPEDANCE AUDIT          ")
	println("=======================================================")
	println("Auditing pins with PullUp (13k) vs PullDown (13k).")
	println("An unconnected floating pin follows pull direction (PU=1, PD=0).")
	println("A grounded load (buzzer/coil) forces PU=0.")
	println("A powered load (pullup/active module) forces PD=1.")
	println("-------------------------------------------------------")

	board.EnableVcc()
	time.Sleep(200 * time.Millisecond)

	for _, p := range allHeaderPins {
		// Test with PullUp
		p.pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
		time.Sleep(5 * time.Millisecond)
		valPU := p.pin.Get()

		// Test with PullDown
		p.pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
		time.Sleep(5 * time.Millisecond)
		valPD := p.pin.Get()

		// Test floating
		p.pin.Configure(machine.PinConfig{Mode: machine.PinInput})
		time.Sleep(5 * time.Millisecond)
		valFloat := p.pin.Get()

		classification := "OPEN / FLOATING"
		if !valPU && !valPD {
			classification = "** TIED TO GND / LOW IMPEDANCE LOAD **"
		} else if valPU && valPD {
			classification = "** TIED TO VCC / ACTIVE HIGH MODULE **"
		} else if valPU != true || valPD != false {
			classification = "** ANOMALY / WEAK LOAD **"
		}

		println("Pin", p.name, "| PU:", valPU, "| PD:", valPD, "| Float:", valFloat, "->", classification)
	}

	println("=======================================================")
	println("Audit complete. Monitoring continues...")
	for {
		time.Sleep(time.Hour)
	}
}
