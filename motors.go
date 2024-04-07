package lego

import (
	"os"
	"time"

	m "github.com/IgneousRed/gomisc"
)

const motorPath = "/sys/class/tacho-motor"
const CountPerRot = 360

type MotorType u8

const (
	MTNone MotorType = iota
	MTLarge
	MTMedium
)

type PidK u8

const (
	PKp PidK = iota
	PKi
	PKd
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

// Removes "/n" at the end
func removeNL(b []u8) string {
	return string(b[:len(b)-1])
}

// Read motor attribute
func (mt Motor) read(attribute string) string {
	val, err := os.ReadFile(mt.path + attribute)
	m.FatalErr("", err)
	return removeNL(val)
}

// Write motor attribute
func (mt Motor) write(attribute, value string) {
	m.FatalErr("", os.WriteFile(mt.path+attribute, []u8(value), 0))
}

// Blocks until requested Motors are correctly connected.
func MotorsInit(portA, portB, portC, portD MotorType, dsp Display,
) (a, b, c, d Motor) {
	m.FatalIf(motorsInitialized, "Motors already Initialized")
	ports := [4]MotorType{portA, portB, portC, portD}

	// Begin
	var result [4]Motor
	for {
		// Find connected Motors
		f, err := os.Open(motorPath)
		m.FatalErr("", err)
		names, err := f.Readdirnames(0)
		m.FatalErr("", err)
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
				return removeNL(val)
			}
			port, motor := attribute("address"), attribute("driver_name")
			if retry {
				break
			}
			portNum := map[string]int{
				"ev3-ports:outA": 0,
				"ev3-ports:outB": 1,
				"ev3-ports:outC": 2,
				"ev3-ports:outD": 3,
			}
			motorType := map[string]MotorType{
				"lego-ev3-l-motor": MTLarge,
				"lego-ev3-m-motor": MTMedium,
			}
			i := portNum[port]
			mt, ok := motorType[motor]
			connected[i] = m.Ternary(ok, mt, MTNone)
			result[i].path = path
		}
		if retry {
			continue
		}

		// Does mismatch exist
		mismatch := false
		for i := range result {
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
		dsp.Write(2, 1, "Motors")
		for i := range result {
			dsp.Write(0, i+2, string(rune('A'+i))+": "+
				m.Ternary(connected[i] != ports[i],
					[4]string{"None", "Large", "Medium"}[ports[i]], "Ok",
				),
			)
		}
		time.Sleep(time.Millisecond * 100)
	}

	// Cache MaxSpeed
	for i := range result {
		if ports[i] != MTNone {
			result[i].maxSpeed, _ = m.AToI[uint](result[i].read("max_speed"))
		}
	}
	motorsInitialized = true
	return result[0], result[1], result[2], result[3]
}

// At start all motors are reset
// hold holds the current pos
// hold pid setting effective immediately. Check every value dep
// wrong dutycycle gives error

type MotorCommand u8

const (
	MCRunForever MotorCommand = iota
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
// CRunAbsPos: Runs to PositionTarget then applies StopAction.
// CRunRelPos: Runs additional PositionTarget then applies StopAction.
// CRunTimed: Runs for TimeTarget then applies StopAction.
// CRunDirect: Runs forever using the PowerTarget.
// Unlike other commands, changing PowerTarget takes effect immediately.
// CStop: Applies StopAction.
// CReset: Resets all attributes to default and puts it into Coast mode.
func (mt Motor) Command(command MotorCommand) {
	mt.write("command", motorCommand[command])
}

// Returns the current Power in percentage[-100, 100].
func (mt Motor) Power() (percent int) {
	percent, _ = m.AToI[int](mt.read("duty_cycle"))
	return
}

// Returns the PowerTarget in percentage[-100, 100].
func (mt Motor) PowerTarget() (percent int) {
	percent, _ = m.AToI[int](mt.read("duty_cycle_sp"))
	return
}

// Declares the PowerTarget in percentage[-100, 100].
// Setting outside [-100, 100] range crashes the program.
func (mt Motor) PowerTargetSet(percent int) { // TODO: Test
	mt.write("duty_cycle_sp", m.IToA(percent))
}

// Returns the PID controller constant for holding position.
func (mt Motor) HoldPid(k PidK) (gain uint) {
	gain, _ = m.AToI[uint](mt.read("hold_pid/" + pidK[k]))
	return
}

// Declares the PID controller constant for holding position.
func (mt Motor) HoldPidSet(k PidK, gain uint) {
	mt.write("hold_pid/"+pidK[k], m.IToA(gain))
}

// Returns the theoretical maximum speed given no load.
func (mt Motor) MaxSpeed() (speed uint) {
	return mt.maxSpeed
}

type Polarity u8

const (
	PNormal Polarity = iota
	PInversed
)

var polarityEnum = map[string]Polarity{
	"normal":   PNormal,
	"inversed": PInversed,
}

// Returns the direction of the motor.
func (mt Motor) Polarity() Polarity {
	return polarityEnum[mt.read("polarity")]
}

var polarityString = [2]string{
	"normal",
	"inversed",
}

// Declares the direction of the motor.
// Changing inverts values for position, speed and dutycycle
func (mt Motor) PolaritySet(polarity Polarity) {
	mt.write("polarity", polarityString[polarity])
}

// Returns the current tacho count.
func (mt Motor) Position() (count int) {
	count, _ = m.AToI[int](mt.read("position"))
	return
}

// Declares the current tacho count.
func (mt Motor) PositionSet(count int) {
	mt.write("position", m.IToA(count))
}

// Returns the PositionTarget.
func (mt Motor) PositionTarget() (count int) {
	count, _ = m.AToI[int](mt.read("position_sp"))
	return
}

// Declares the PositionTarget.
func (mt Motor) PositionTargetSet(count int) {
	mt.write("position_sp", m.IToA(count))
}

// Returns the RampDown time in millis.
// RampDown is time it takes the motor to decrease Power
// from 100% to 0% in X modes. TODO: Test
func (mt Motor) RampDown() (millis uint) {
	millis, _ = m.AToI[uint](mt.read("ramp_down_sp"))
	return
}

// Declares the RampDown time in millis.
// RampDown is time it takes the motor to decrease Power
// from 100% to 0% in X modes.
func (mt Motor) RampDownSet(millis uint) {
	mt.write("ramp_down_sp", m.IToA(millis))
}

// Returns the RampUp time in millis.
// RampUp is time it takes the motor to increase Power
// from 0% to 100% in X modes.
func (mt Motor) RampUp() (millis uint) {
	millis, _ = m.AToI[uint](mt.read("ramp_up_sp"))
	return
}

// Declares the RampUp time in millis.
// RampUp is time it takes the motor to increase Power
// from 0% to 100% in X modes.
func (mt Motor) RampUpSet(millis uint) {
	mt.write("ramp_up_sp", m.IToA(millis))
}

// Returns the current motor speed.
func (mt Motor) Speed() (speed int) {
	speed, _ = m.AToI[int](mt.read("speed"))
	return
}

// Returns the PID controller constant for maintaining speed.
func (mt Motor) SpeedPid(k PidK) (gain uint) {
	gain, _ = m.AToI[uint](mt.read("speed_pid/" + pidK[k]))
	return
}

// Declares the PID controller constant for maintaining speed.
func (mt Motor) SpeedPidSet(k PidK, gain uint) {
	mt.write("speed_pid/"+pidK[k], m.IToA(gain))
}

// Returns the SpeedTarget.
func (mt Motor) SpeedTarget() (speed int) {
	speed, _ = m.AToI[int](mt.read("speed_sp"))
	return
}

// Declares the SpeedTarget.
// Setting outside [-MaxSpeed, MaxSpeed] range crashes the program.
func (mt Motor) SpeedTargetSet(speed int) {
	mt.write("speed_sp", m.IToA(speed))
}

type State u8

const (
	SNil State = iota
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
func (mt Motor) State() State {
	return stateEnum[mt.read("state")]
}

type StopAction u8

const (
	SACoast StopAction = iota
	SABrake
	SAHold
)

var sAEnum = map[string]StopAction{
	"coast": SACoast,
	"brake": SABrake,
	"hold":  SAHold,
}

// TODO: Desc
func (mt Motor) StopAction() StopAction {
	return sAEnum[mt.read("stop_action")]
}

var sAString = [3]string{
	"coast",
	"brake",
	"hold",
}

// TODO: Desc
func (mt Motor) StopActionSet(action StopAction) {
	mt.write("stop_action", sAString[action])
}

// Returns the TimeTarget in millis.
func (mt Motor) TimeTarget() (millis uint) {
	millis, _ = m.AToI[uint](mt.read("time_sp"))
	return
}

// Declares the TimeTarget in millis.
func (mt Motor) TimeTargetSet(millis uint) {
	mt.write("time_sp", m.IToA(millis))
}
