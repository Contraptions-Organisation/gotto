
package main

import (
	"machine"
	"time"

	"github.com/HattoriHanzo031/gotto/buzzer"
)

func main() {
	time.Sleep(2 * time.Second)
	println("Testing PWM0 on P1.13...")
	bz := buzzer.New(buzzer.NewPwmChannel(machine.PWM0, machine.P1_13))
	if err := bz.Configure(); err != nil {
		println("Config error:", err.Error())
		return
	}
	println("Configured! Playing A4...")
	for i := 0; i < 5; i++ {
		err := bz.Tone(buzzer.Note{Period: buzzer.A4, Duration: 500 * time.Millisecond})
		if err != nil {
			println("Tone error:", err.Error())
		}
		time.Sleep(300 * time.Millisecond)
	}
	println("Done!")
}
