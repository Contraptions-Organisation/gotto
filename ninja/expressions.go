package ninja

import (
	"time"

	"github.com/HattoriHanzo031/gotto/board"
	"github.com/HattoriHanzo031/gotto/buzzer"
)

// HappyWiggle performs an energetic happy wiggle with a cheerful rising arpeggio.
func (n *Ninja) HappyWiggle() error {
	if n.mode != ModeWalk {
		return ErrInvalidMode
	}

	// Cheerful arpeggio melody
	if n.buzzer != nil {
		go func() {
			n.buzzer.PlayMelody([]buzzer.Note{
				{Period: buzzer.C5, Duration: 70 * time.Millisecond},
				{Period: buzzer.E5, Duration: 70 * time.Millisecond},
				{Period: buzzer.G5s, Duration: 70 * time.Millisecond},
				{Period: buzzer.C6, Duration: 180 * time.Millisecond},
			})
		}()
	}

	// Rapid playful side-to-side body shake
	for range 3 {
		n.lLegAngle(110)
		n.rLegAngle(70)
		time.Sleep(100 * time.Millisecond)
		n.lLegAngle(70)
		n.rLegAngle(110)
		time.Sleep(100 * time.Millisecond)
	}

	return n.Home()
}

// Moonwalk performs Michael Jackson-style backward sliding steps.
func (n *Ninja) Moonwalk(steps int) error {
	if n.mode != ModeWalk {
		return ErrInvalidMode
	}

	for range steps {
		// Tilt left, slide right foot backward smoothly
		n.Tilt(TiltLeft)
		n.MoveRightFoot(-35, 450*time.Millisecond)
		n.Tilt(TiltReturnFromLeft)

		// Tilt right, slide left foot backward smoothly
		n.Tilt(TiltRight)
		n.MoveLeftFoot(-35, 450*time.Millisecond)
		n.Tilt(TiltReturnFromRight)
	}

	return n.Home()
}

// PivotTurn turns the robot in place on one foot.
// Positive turnAngle turns right, negative turns left.
func (n *Ninja) PivotTurn(turns int) error {
	if n.mode != ModeWalk {
		return ErrInvalidMode
	}

	if turns > 0 {
		// Turn right by pivoting on left foot
		for range turns {
			n.Tilt(TiltLeft)
			n.MoveLeftFoot(45, 500*time.Millisecond)
			n.Tilt(TiltReturnFromLeft)
		}
	} else {
		// Turn left by pivoting on right foot
		for range -turns {
			n.Tilt(TiltRight)
			n.MoveRightFoot(-45, 500*time.Millisecond)
			n.Tilt(TiltReturnFromRight)
		}
	}

	return n.Home()
}

// CuriousLook tilts inquiringly to one side and plays an interrogative chirp.
func (n *Ninja) CuriousLook() error {
	if n.mode != ModeWalk {
		return ErrInvalidMode
	}

	if n.buzzer != nil {
		go func() {
			n.buzzer.PlayMelody([]buzzer.Note{
				{Period: buzzer.E5, Duration: 120 * time.Millisecond},
				{Period: buzzer.Silence, Duration: 60 * time.Millisecond},
				{Period: buzzer.G5s, Duration: 240 * time.Millisecond},
			})
		}()
	}

	n.lLegAngle(115)
	n.rLegAngle(85)
	time.Sleep(700 * time.Millisecond)

	return n.Home()
}

// YawnAndSleep plays a gentle descending lullaby, bows forward, and cuts
// the switched VCC servo rail to eliminate servo hum and save battery.
func (n *Ninja) YawnAndSleep() error {
	if n.buzzer != nil {
		n.buzzer.PlayMelody([]buzzer.Note{
			{Period: buzzer.G5s, Duration: 200 * time.Millisecond},
			{Period: buzzer.E5, Duration: 250 * time.Millisecond},
			{Period: buzzer.C5, Duration: 350 * time.Millisecond},
			{Period: buzzer.C4, Duration: 500 * time.Millisecond},
		})
	}

	// Home and cut power to servos
	_ = n.Home()
	time.Sleep(300 * time.Millisecond)

	// Disabling switched VCC rail cuts power to all 4 servos
	board.DisableVcc()
	return nil
}

// WakeUp re-enables the switched VCC rail, centers posture, and plays a wake-up fanfare.
func (n *Ninja) WakeUp() error {
	board.EnableVcc()
	time.Sleep(200 * time.Millisecond)

	if n.buzzer != nil {
		go func() {
			n.buzzer.PlayMelody([]buzzer.Note{
				{Period: buzzer.C4, Duration: 100 * time.Millisecond},
				{Period: buzzer.E4, Duration: 100 * time.Millisecond},
				{Period: buzzer.G4, Duration: 100 * time.Millisecond},
				{Period: buzzer.C5, Duration: 250 * time.Millisecond},
			})
		}()
	}

	return n.Home()
}
