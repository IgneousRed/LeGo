package main

import (
	"log"
	"os"
	"time"

	misc "github.com/IgneousRed/gomisc"
)

const motorPath = "/sys/class/tacho-motor"
const CountPerRot = 360

type MotorType byte

const (
	MTNone = MotorType(iota)
	MTLarge
	MTMedium
)

type PidK byte

const (
	PKP = PidK(iota)
	PKI
	PKD
)

var pidK = [3]string{
	"Kp",
	"Ki",
	"Kd",
}

type Motor struct {
	path     string
	maxSpeed uint
}

var motorsInitialized bool

func (m Motor) read(attribute string) string {
	val, err := os.ReadFile(m.path + attribute)
	misc.FatalErr("", err)
	return string(val[:len(val)-1])
}
func (m Motor) write(attribute, value string) {
	f, err := os.OpenFile(m.path+attribute, os.O_WRONLY|os.O_TRUNC, 0)
	misc.FatalErr("", err)
	_, err = f.WriteString(value)
	misc.FatalErr("", err)
	f.Close()
}

// Blocks until requested Motors are correctly connected.
// Initialize Display before Motors.
func MotorsInit(portA, portB, portC, portD MotorType, dsp Display) (a, b, c, d Motor) {
	misc.FatalIf(motorsInitialized, "Motors already Initialized")
	ports := [4]MotorType{portA, portB, portC, portD}
	var motors [4]Motor
	for {
		// Find connected Motors
		f, err := os.Open(motorPath)
		misc.FatalErr("", err)
		names, err := f.Readdirnames(0)
		misc.FatalErr("", err)
		f.Close()

		// Populate conections
		connected := [4]MotorType{}
		retry := false // If a Motor disconnects, restart
		for _, name := range names {
			path := motorPath + "/" + name + "/"
			attribute := func(att string) string {
				val, err := os.ReadFile(path + att)
				if err != nil {
					retry = true
					return ""
				}
				return string(val[:len(val)-1]) // Removes "\n" at the end
			}
			port := attribute("address")
			motor := attribute("driver_name")
			if retry {
				break
			}
			var portNum = map[string]int{
				"ev3-ports:outA": 0,
				"ev3-ports:outB": 1,
				"ev3-ports:outC": 2,
				"ev3-ports:outD": 3,
			}
			var motorType = map[string]MotorType{ // warning on other types
				"lego-ev3-l-motor": MTLarge,
				"lego-ev3-m-motor": MTMedium,
			}
			i := portNum[port]
			connected[i] = motorType[motor]
			motors[i].path = path
		}
		if retry {
			continue
		}

		// Does mismatch exist
		mismatch := false
		for i := range motors {
			if connected[i] != ports[i] {
				mismatch = true
				break
			}
		}
		dsp.Clear()
		if !mismatch {
			break
		}

		// Prints help
		dsp.Write(1, 2, "Motors")
		for i := range motors {
			var motorName = [3]string{"None", "Large", "Medium"}
			postFix := misc.Ternary(connected[i] != ports[i], motorName[ports[i]], "Ok")
			dsp.Write(i+2, 0, string(rune('A'+i))+": "+postFix)
		}
		time.Sleep(time.Millisecond * 100)
	}

	// Cache MaxSpeed
	for i := range motors {
		if ports[i] != MTNone {
			max, _ := misc.AToI[uint](motors[i].read("max_speed"))
			motors[i].maxSpeed = max
		}
	}
	motorsInitialized = true
	return motors[0], motors[1], motors[2], motors[3]
}

// At start all motors are reset
// hold holds the current pos
// hold pid setting effective immediately. Check every value dep
// wrong dutycycle gives error

type MotorCommand byte

const (
	MCRunForever = MotorCommand(iota)
	MCRunAbsPos
	MCRunRelPos
	MCRunTimed
	MCRunDirect
	MCStop
	MCReset
)

var motorCommand = [7]string{
	"run-forever",
	"run-to-abs-pos",
	"run-to-rel-pos",
	"run-timed",
	"run-direct",
	"stop",
	"reset",
}

// CRunForever: Runs forever.
// CRunAbsPos: Runs to PositionSetpoint then applies StopAction.
// CRunRelPos: Runs additional PositionSetpoint then applies StopAction.
// CRunTimed: Runs for TimeSetpoint then applies StopAction.
// CRunDirect: Runs forever using the DutyCycleSetpoint.
// Unlike other commands, changing DutyCycleSetpoint takes effect immediately.
// CStop: Applies StopAction.
// CReset: Resets all attributes to default and puts it into Coast mode.
func (m Motor) Command(command MotorCommand) {
	m.write("command", motorCommand[command])
}

// Returns the current DutyCycle(power given) in percentage[-100, 100].
func (m Motor) DutyCycle() (percent int) {
	percent, _ = misc.AToI[int](m.read("duty_cycle"))
	return
}

// Returns the DutyCycleSetpoint in percentage[-100, 100].
func (m Motor) DutyCycleSetpointGet() (percent int) {
	percent, _ = misc.AToI[int](m.read("duty_cycle_sp"))
	return
}

// Declares the DutyCycleSetpoint in percentage[-100, 100].
func (m Motor) DutyCycleSetpointSet(percent int) {
	clamped, flag := misc.ClampReport(percent, -100, 100)
	if flag {
		log.Println("Warning: DutyCycle " + misc.IToA(percent) + " is outside [-100, 100] range")
	}
	m.write("duty_cycle_sp", misc.IToA(clamped))
}

// Returns the PID controller constant for holding position.
func (m Motor) HoldPidGet(k PidK) (gain uint) {
	gain, _ = misc.AToI[uint](m.read("hold_pid/" + pidK[k]))
	return
}

// Declares the PID controller constant for holding position.
func (m Motor) HoldPidSet(k PidK, gain uint) {
	m.write("hold_pid/"+pidK[k], misc.IToA(gain))
}

// Returns the theoretical maximum speed given no load.
func (m Motor) MaxSpeed() (speed uint) {
	return m.maxSpeed
}

type Polarity byte

const (
	PNormal = Polarity(iota)
	PInversed
)

var polarityEnum = map[string]Polarity{
	"normal":   PNormal,
	"inversed": PInversed,
}

// Returns the direction of the motor.
func (m Motor) PolarityGet() Polarity {
	return polarityEnum[m.read("polarity")]
}

var polarityString = [2]string{
	"normal",
	"inversed",
}

// Declares the direction of the motor.
// Changing inverts values for position, speed and dutycycle
func (m Motor) PolaritySet(polarity Polarity) {
	m.write("polarity", polarityString[polarity])
}

// Returns the current tacho count.
func (m Motor) PositionGet() (count int) {
	count, _ = misc.AToI[int](m.read("position"))
	return
}

// Declares the current tacho count.
func (m Motor) PositionSet(count int) {
	m.write("position", misc.IToA(count))
}

// Returns the PositionSetpoint.
func (m Motor) PositionSetpointGet() (count int) {
	count, _ = misc.AToI[int](m.read("position_sp"))
	return
}

// Declares the PositionSetpoint.
func (m Motor) PositionSetpointSet(count int) {
	m.write("position_sp", misc.IToA(count))
}

// Returns the RampDown time in millis.
// RampDown is time it takes the motor to decrease DutyCycle from 100% to 0% in X modes. TODO: Test
func (m Motor) RampDownGet() (millis uint) {
	millis, _ = misc.AToI[uint](m.read("ramp_down_sp"))
	return
}

// Declares the RampDown time in millis.
// RampDown is time it takes the motor to decrease DutyCycle from 100% to 0% in X modes.
func (m Motor) RampDownSet(millis uint) {
	m.write("ramp_down_sp", misc.IToA(millis))
}

// Returns the RampUp time in millis.
// RampUp is time it takes the motor to increase DutyCycle from 0% to 100% in X modes.
func (m Motor) RampUpGet() (millis uint) {
	millis, _ = misc.AToI[uint](m.read("ramp_up_sp"))
	return
}

// Declares the RampUp time in millis.
// RampUp is time it takes the motor to increase DutyCycle from 0% to 100% in X modes.
func (m Motor) RampUpSet(millis uint) {
	m.write("ramp_up_sp", misc.IToA(millis))
}

// Returns the current motor speed.
func (m Motor) Speed() (speed int) {
	speed, _ = misc.AToI[int](m.read("speed"))
	return
}

// Returns the PID controller constant for maintaining speed.
func (m Motor) SpeedPidGet(k PidK) (gain uint) {
	gain, _ = misc.AToI[uint](m.read("speed_pid/" + pidK[k]))
	return
}

// Declares the PID controller constant for maintaining speed.
func (m Motor) SpeedPidSet(k PidK, gain uint) {
	m.write("speed_pid/"+pidK[k], misc.IToA(gain))
}

// Returns the SpeedSetpoint.
func (m Motor) SpeedSetpointGet() (speed int) {
	speed, _ = misc.AToI[int](m.read("speed_sp"))
	return
}

// Declares the SpeedSetpoint.
func (m Motor) SpeedSetpointSet(speed int) {
	m.write("speed_sp", misc.IToA(speed))
}

type State byte

const (
	SNil = State(iota)
	SRunning
	SRamping
	SHolding
	SOverloaded
	SStalled
)

var stateEnum = map[string]State{
	"":           SNil,
	"running":    SRunning,
	"ramping":    SRamping,
	"holding":    SHolding,
	"overloaded": SOverloaded,
	"stalled":    SStalled,
}

// TODO: Desc
func (m Motor) State() State {
	return stateEnum[m.read("state")]
}

type StopAction byte

const (
	SACoast = StopAction(iota)
	SABrake
	SAHold
)

var sAEnum = map[string]StopAction{
	"coast": SACoast,
	"brake": SABrake,
	"hold":  SAHold,
}

// TODO: Desc
func (m Motor) StopActionGet() StopAction {
	return sAEnum[m.read("stop_action")]
}

var sAString = [3]string{
	"coast",
	"brake",
	"hold",
}

// TODO: Desc
func (m Motor) StopActionSet(action StopAction) {
	m.write("stop_action", sAString[action])
}

// Returns the TimeSetpoint in millis.
func (m Motor) TimeSetpointGet() (millis uint) {
	millis, _ = misc.AToI[uint](m.read("time_sp"))
	return
}

// Declares the TimeSetpoint in millis.
func (m Motor) TimeSetpointSet(millis uint) {
	m.write("time_sp", misc.IToA(millis))
}
