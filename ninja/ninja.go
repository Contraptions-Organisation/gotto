package ninja

import (
	"errors"
	"time"

	"github.com/HattoriHanzo031/gotto/buzzer"
	"github.com/HattoriHanzo031/gotto/servo"
)

const (
	walkSpeed         = 50
	tiltAngle         = 45
	stepDuration      = 600 * time.Millisecond
	numCustomCommands = 10
)

// Mode represents the mode of the robot, which can be either walk or roll.
type Mode int

const (
	ModeWalk Mode = iota
	ModeRoll
)

// TiltDir represents the direction for tilting (or returning from tilt) the robot.
type TiltDir int

const (
	TiltReturnFromLeft TiltDir = iota
	TiltReturnFromRight
	TiltLeft
	TiltRight
)

// CustomCommand represents a user-defined command that can be executed on the Ninja robot.
type CustomCommand func(*Ninja) error

var (
	ErrInvalidMode                  = errors.New("ninja: invalid mode")
	ErrInvalidDirection             = errors.New("ninja: invalid direction")
	ErrBuzzerNotConfigured          = errors.New("ninja: buzzer not configured")
	ErrCustomCommandIndexOutOfRange = errors.New("ninja: custom command index out of range")
)

// Trim represents the trim values for the robot's movement and posture adjustments.
// TODO: trim should be persisted in non-volatile memory and applied on startup.
type Trim struct {
	// TiltAngle is added to the tilt angle when tilting.
	// Positive values increase the tilt angle, while negative values decrease it.
	TiltAngle int
	// LeftStepDuration is added to the duration of each left step when walking.
	// Positive values make the left steps longer, while negative values make them shorter.
	LeftStepDuration time.Duration
	// RightStepDuration is added to the duration of each right step when walking.
	// Positive values make the right steps longer, while negative values make them shorter.
	RightStepDuration time.Duration
	// LfSpeed is added to the speed of the left foot when moving.
	// Positive values make the left foot faster, while negative values make it slower.
	LfSpeed int
	// RfSpeed is added to the speed of the right foot when moving.
	// Positive values make the right foot faster, while negative values make it slower.
	RfSpeed int
	// LlAngle is added to the angle of the left leg when moving.
	// Positive values increase the angle, while negative values decrease it.
	LlAngle int
	// RlAngle is added to the angle of the right leg when moving.
	// Positive values increase the angle, while negative values decrease it.
	RlAngle int
}

type Ninja struct {
	rLeg           servo.Servo180
	rFoot          servo.Servo360
	lLeg           servo.Servo180
	lFoot          servo.Servo360
	llAngle        int
	rlAngle        int
	mode           Mode
	err            error
	trim           Trim
	buzzer         *buzzer.Buzzer
	customCommands [numCustomCommands]CustomCommand
}

// New creates a new Ninja instance with the given leg and foot servos and an optional buzzer.
// The leg servos should be of type Servo180, while the foot servos should be of type Servo360.
// The buzzer can be nil if not used, but it is required for using the BuzzerTone method.
// The servos should be configured and ready to use before creating the Ninja instance.
func New(rLeg, lLeg servo.Servo180, rFoot, lFoot servo.Servo360, buzzer *buzzer.Buzzer) *Ninja {
	return &Ninja{
		rLeg:    rLeg,
		rFoot:   rFoot,
		lLeg:    lLeg,
		lFoot:   lFoot,
		llAngle: 95,
		rlAngle: 95,
		trim:    Trim{},
		mode:    ModeWalk,
		buzzer:  buzzer,
	}
}

// setAngleSmooth gradually changes the angle from current to new in 30 steps
// TODO: make step count and delay configurable
func setAngleSmooth(new, current int, set func(int) error) error {
	increment := float32(new-current) / 30.0
	for i := range 30 {
		if err := set(current + int(increment*float32(i+1))); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
	}
	return nil
}

func (n *Ninja) lLegAngle(angle int) {
	if n.err != nil {
		return
	}

	angle += n.trim.LlAngle
	angle = 180 - angle

	n.err = setAngleSmooth(angle, n.llAngle, n.lLeg.SetAngle)
	if n.err != nil {
		return
	}
	n.llAngle = angle
}

func (n *Ninja) rLegAngle(angle int) {
	if n.err != nil {
		return
	}

	angle += n.trim.RlAngle

	n.err = setAngleSmooth(angle, n.rlAngle, n.rLeg.SetAngle)
	if n.err != nil {
		return
	}
	n.rlAngle = angle
}

func speedTrim(speed, trim int) int {
	switch {
	case speed > 0:
		speed = min(speed+trim, 100)
	case speed < 0:
		speed = max(speed-trim, -100)
	}
	return speed
}

func (n *Ninja) rFootSpeed(speed int) {
	if n.err != nil {
		return
	}

	speed = speedTrim(-speed, n.trim.RfSpeed)
	n.err = n.rFoot.SetSpeed(speed)
}

func (n *Ninja) lFootSpeed(speed int) {
	if n.err != nil {
		return
	}

	speed = speedTrim(speed, n.trim.LfSpeed)
	n.err = n.lFoot.SetSpeed(speed)
}

func (n *Ninja) error() error {
	err := n.err
	n.err = nil
	return err
}

// Trim sets the trim values for the robot.
// Trim can be used to adjust the robot's movement if it's not moving straight or
// if the legs are not at the same angle in the home position.
func (n *Ninja) Trim(trim Trim) {
	n.trim = trim
}

// Tilt performs a tilting motion in the specified direction.
// dir can be TiltLeft, TiltRight, TiltReturnFromLeft, or TiltReturnFromRight.
// It requires the robot to be in walk mode.
func (n *Ninja) Tilt(dir TiltDir) error {
	if n.mode != ModeWalk {
		return ErrInvalidMode
	}
	angle := tiltAngle + n.trim.TiltAngle
	switch dir {
	case TiltReturnFromLeft:
		n.lLegAngle(90)
		n.rLegAngle(90)
	case TiltReturnFromRight:
		n.rLegAngle(90)
		n.lLegAngle(90)
	case TiltLeft:
		n.rLegAngle(90 + angle + 15)
		n.lLegAngle(90 - angle)
	case TiltRight:
		n.lLegAngle(90 + angle + 15)
		n.rLegAngle(90 - angle)
	default:
		return ErrInvalidDirection
	}

	return n.error()
}

// Mode sets the robot's mode to either walk or roll.
// It also moves the robot to its home position for the new mode.
func (n *Ninja) Mode(mode Mode) error {
	if n.mode != ModeWalk && mode != ModeRoll {
		return ErrInvalidMode
	}

	n.mode = mode
	return n.Home()
}

// Home moves the robot to its home position.
// Home position in walk mode is standing straight with feet together.
// Home position in roll mode both legs raised to the side.
func (n *Ninja) Home() error {
	n.lFootSpeed(0)
	n.rFootSpeed(0)
	switch n.mode {
	case ModeWalk:
		n.lLegAngle(90)
		n.rLegAngle(90)
	case ModeRoll:
		n.lLegAngle(0)
		time.Sleep(200 * time.Millisecond)
		n.rLegAngle(0)
	}
	return n.error()
}

// MoveLeftFoot spins the left foot with the given speed and duration, then stops it.
func (n *Ninja) MoveLeftFoot(speed int, duration time.Duration) error {
	n.lFootSpeed(speed)
	time.Sleep(duration)
	n.lFootSpeed(0)
	return n.error()
}

// MoveRightFoot spins the right foot with the given speed and duration, then stops it.
func (n *Ninja) MoveRightFoot(speed int, duration time.Duration) error {
	n.rFootSpeed(speed)
	time.Sleep(duration)
	n.rFootSpeed(0)
	return n.error()
}

// LeftLegSpin tilts to the left (planting the left foot on the ground) and spins the left foot forward.
func (n *Ninja) LeftLegSpin(speed int, duration time.Duration) error {
	if n.mode != ModeWalk {
		return ErrInvalidMode
	}
	n.Tilt(TiltLeft)
	n.MoveLeftFoot(speed, duration)
	n.Tilt(TiltReturnFromLeft)
	return n.error()
}

// RightLegSpin tilts to the right (planting the right foot on the ground) and spins the right foot forward.
func (n *Ninja) RightLegSpin(speed int, duration time.Duration) error {
	if n.mode != ModeWalk {
		return ErrInvalidMode
	}
	n.Tilt(TiltRight)
	n.MoveRightFoot(speed, duration)
	n.Tilt(TiltReturnFromRight)
	return n.error()
}

// Walk performs a walking motion for the given number of steps.
// Positive steps walk forward, while negative steps walk backward.
// Each step consists of stepping with both legs.
// It requires the robot to be in walk mode.
func (n *Ninja) Walk(steps int) error {
	if n.mode != ModeWalk {
		return ErrInvalidMode
	}

	if steps == 0 {
		return nil
	}

	speed := walkSpeed
	if steps < 0 {
		speed = -walkSpeed
		steps = -steps
	}

	rStepDuration := stepDuration + n.trim.RightStepDuration
	lStepDuration := stepDuration + n.trim.LeftStepDuration

	for range steps {
		n.err = n.LeftLegSpin(speed, lStepDuration)
		n.err = n.RightLegSpin(speed, rStepDuration)
	}

	return n.error()
}

// SetCustomCommand sets a custom command function at the given index.
// The custom command can be executed later by sending a command with OpCustom and the same index.
// If the index is out of range, it returns an ErrCustomCommandIndexOutOfRange error.
func (n *Ninja) SetCustomCommand(index int, fn CustomCommand) error {
	if index < 0 || index >= len(n.customCommands) {
		return ErrCustomCommandIndexOutOfRange
	}
	n.customCommands[index] = fn
	return nil
}

// ExecuteCustomCommand executes the custom command at the given index.
// If the index is out of range, it returns an ErrCustomCommandIndexOutOfRange error.
// If there is no custom command set at the index, the command has no effect.
func (n *Ninja) ExecuteCustomCommand(index int) error {
	if index < 0 || index >= len(n.customCommands) {
		return ErrCustomCommandIndexOutOfRange
	}
	if n.customCommands[index] == nil {
		return nil
	}

	return n.customCommands[index](n)
}

// Roll performs a rolling motion with the given throttle and turn values.
// Throttle controls the forward/backward speed, while turn controls the turning speed.
// Throttle and turn should be in the range -100 to 100.
// Positive turn values turn right, while negative values turn left.
// It requires the robot to be in roll mode.
func (n *Ninja) Roll(throttle, turn int) error {
	if n.mode != ModeRoll {
		return ErrInvalidMode
	}

	n.lFootSpeed(throttle + turn)
	n.rFootSpeed(throttle - turn)
	return n.error()
}

func (n *Ninja) RollStop() error {
	return n.Roll(0, 0)
}

// StartLeftSpin starts spinning the robot on left leg.
// Robot spins until StopLeftSpin is called.
// It requires the robot to be in walk mode.
func (n *Ninja) StartLeftSpin(speed int) error {
	if n.mode != ModeWalk {
		return ErrInvalidMode
	}

	n.err = n.Tilt(TiltLeft)
	n.lFootSpeed(speed)
	return n.error()
}

// StopLeftSpin stops the left leg spin started by StartLeftSpin.
func (n *Ninja) StopLeftSpin() error {
	if n.mode != ModeWalk {
		return ErrInvalidMode
	}
	n.lFootSpeed(0)
	n.err = n.Tilt(TiltReturnFromLeft)
	return n.error()
}

// StartRightSpin starts spinning the robot on right leg.
// Robot spins until StopRightSpin is called.
// It requires the robot to be in walk mode.
func (n *Ninja) StartRightSpin(speed int) error {
	if n.mode != ModeWalk {
		return ErrInvalidMode
	}

	n.err = n.Tilt(TiltRight)
	n.rFootSpeed(speed)
	return n.error()
}

// StopRightSpin stops the spinning started by StartRightSpin.
func (n *Ninja) StopRightSpin() error {
	if n.mode != ModeWalk {
		return ErrInvalidMode
	}
	n.rFootSpeed(0)
	n.err = n.Tilt(TiltReturnFromRight)
	return n.error()
}

// Wave performs a waving motion with the left leg. It requires the robot to be in walk mode.
func (n *Ninja) Wave() error {
	if n.mode != ModeWalk {
		return ErrInvalidMode
	}
	n.err = n.Tilt(TiltRight)
	time.Sleep(500 * time.Millisecond)

	angle := n.llAngle
	for range 4 {
		n.lLegAngle(angle + 30)
		n.lLegAngle(angle)
	}

	time.Sleep(500 * time.Millisecond)
	n.err = n.Tilt(TiltReturnFromRight)
	return n.error()
}

// BuzzerTone plays a tone on the buzzer
// If  buzzer is not configured, it returns an ErrBuzzerNotConfigured error
func (n *Ninja) BuzzerTone(note buzzer.Note) error {
	if n.buzzer == nil {
		return ErrBuzzerNotConfigured
	}
	return n.buzzer.Tone(note)
}
