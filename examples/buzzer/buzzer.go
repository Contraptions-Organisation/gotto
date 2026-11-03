package main

import (
	"machine"
	"time"

	"github.com/HattoriHanzo031/gotto/buzzer"
)

var pwm = machine.PWM0
var buzzerPin = machine.P0_31

var melodyImperial []buzzer.Note = []buzzer.Note{
	{Period: buzzer.A4, Duration: 500 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 200 * time.Millisecond},
	{Period: buzzer.A4, Duration: 500 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 200 * time.Millisecond},
	{Period: buzzer.A4, Duration: 500 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 200 * time.Millisecond},
	{Period: buzzer.F4, Duration: 400 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 50 * time.Millisecond},
	{Period: buzzer.C5, Duration: 200 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 50 * time.Millisecond},
	{Period: buzzer.A4, Duration: 600 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 100 * time.Millisecond},
	{Period: buzzer.F4, Duration: 400 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 50 * time.Millisecond},
	{Period: buzzer.C5, Duration: 200 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 50 * time.Millisecond},
	{Period: buzzer.A4, Duration: 600 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 800 * time.Millisecond},
	{Period: buzzer.E5, Duration: 500 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 200 * time.Millisecond},
	{Period: buzzer.E5, Duration: 500 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 200 * time.Millisecond},
	{Period: buzzer.E5, Duration: 500 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 200 * time.Millisecond},
	{Period: buzzer.F5, Duration: 400 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 50 * time.Millisecond},
	{Period: buzzer.C5, Duration: 200 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 50 * time.Millisecond},
	{Period: buzzer.G5s, Duration: 600 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 100 * time.Millisecond},
	{Period: buzzer.F4, Duration: 400 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 50 * time.Millisecond},
	{Period: buzzer.C5, Duration: 200 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 50 * time.Millisecond},
	{Period: buzzer.A4, Duration: 600 * time.Millisecond},
	{Period: buzzer.Silence, Duration: 400 * time.Millisecond},
}

var melodyTheme = []buzzer.Note{
	{Period: buzzer.F4, Duration: 210 * time.Millisecond},
	{Period: buzzer.F4, Duration: 210 * time.Millisecond},
	{Period: buzzer.F4, Duration: 210 * time.Millisecond},
	{Period: buzzer.A4s, Duration: 1280 * time.Millisecond},
	{Period: buzzer.F5, Duration: 1280 * time.Millisecond},
	{Period: buzzer.D5s, Duration: 210 * time.Millisecond},
	{Period: buzzer.D5, Duration: 210 * time.Millisecond},
	{Period: buzzer.C5, Duration: 210 * time.Millisecond},
	{Period: buzzer.A5s, Duration: 1280 * time.Millisecond},
	{Period: buzzer.F5, Duration: 640 * time.Millisecond},
	{Period: buzzer.D5s, Duration: 210 * time.Millisecond},
	{Period: buzzer.D5, Duration: 210 * time.Millisecond},
	{Period: buzzer.C5, Duration: 210 * time.Millisecond},
	{Period: buzzer.A5s, Duration: 1280 * time.Millisecond},
	{Period: buzzer.F5, Duration: 640 * time.Millisecond},
	{Period: buzzer.D5s, Duration: 210 * time.Millisecond},
	{Period: buzzer.D5, Duration: 210 * time.Millisecond},
	{Period: buzzer.D5s, Duration: 210 * time.Millisecond},
	{Period: buzzer.C5, Duration: 1280 * time.Millisecond},
}

func main() {
	time.Sleep(3 * time.Second)

	buzz := buzzer.New(buzzer.NewPwmChannel(pwm, buzzerPin))
	err := buzz.Configure()
	if err != nil {
		println("Error configuring buzzer:", err.Error())
		return
	}

	for {
		buzz.PlayMelody(melodyImperial)
		time.Sleep(5 * time.Second)
		buzz.PlayMelody(melodyTheme)
		time.Sleep(10 * time.Second)
	}
}
