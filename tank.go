package main

import (
	"fmt"
	"log"
	"time"

	misc "github.com/IgneousRed/gomisc"
)

type Tank struct {
	left, right      Motor
	accK, kp, ki, kd float32
	i, last          float32
	stopPositions    misc.Vec[float32]
	modeHold         bool
}

var tankInitialized bool

func (t *Tank) MovePid(kp, ki, kd float32) {
	t.kp = misc.Pow(2, kp)
	t.ki = misc.Pow(2, ki)
	t.kd = misc.Pow(2, kd)
}
func (t *Tank) Deceleration(value float32) {
	t.accK = value * 2
}
func TankInit(left, right Motor) Tank {
	misc.FatalIf(tankInitialized, "Tank already Initialized")
	var result Tank
	result.left = left
	result.right = right
	result.stopPositions = []float32{0, 0}
	result.modeHold = true
	result.MovePid(-8, -30, -1) // test
	result.Deceleration(999999) // test
	left.StopActionSet(SAHold)
	right.StopActionSet(SAHold)
	left.Command(MCStop)
	right.Command(MCStop)
	return result
}
func (t Tank) positions() misc.Vec[float32] {
	left, right := t.left.PositionGet(), t.right.PositionGet()
	return []float32{float32(left), float32(right)}
}
func (t *Tank) Power(left, right int) {
	if t.modeHold {
		t.left.Command(MCRunDirect)
		t.right.Command(MCRunDirect)
		t.modeHold = false
	}
	t.left.DutyCycleSetpointSet(left)
	t.right.DutyCycleSetpointSet(right)
}
func (t *Tank) Stop() {
	if t.modeHold {
		log.Println("Warning: Tank already stopped")
		return
	}
	t.left.Command(MCStop)
	t.right.Command(MCStop)
	t.modeHold = true
}
func (t Tank) TuneDeceleration(dsp Display) {
	t.left.StopActionSet(SACoast)
	t.right.StopActionSet(SACoast)
	t.Power(100, 100)
	topSpd := 0
	timeStart := time.Now().UnixMicro()
	timeTop := timeStart + 100000
	for {
		timeNow := time.Now().UnixMicro()
		if s := t.left.Speed(); s > topSpd {
			timeTop = timeNow + 100000
			topSpd = s
		}
		if s := t.right.Speed(); s > topSpd {
			timeTop = timeNow + 100000
			topSpd = s
		}
		if timeNow >= timeTop {
			break
		}
	}
	log.Println(timeTop - timeStart)
	t.left.Command(MCStop)
	t.right.Command(MCStop)
	for t.left.Speed() != 0 || t.right.Speed() != 0 {
	}
	// os.WriteFile("/IRLib/Deceleration", )
	log.Println(.0002 * float32(time.Now().UnixMicro()-timeTop))
}
func (t *Tank) TunePID(dsp Display) {
	maxSpd := float32(100)
	cursorLine := 0
	errAcc := float32(0)
	errCount := float32(0)
	lastErrRate := ""
	buttonsOld := Buttons()
	cursor := func(y int) string {
		if y == cursorLine {
			return ">"
		}
		return " "
	}
	modifyInput := func(increase bool) {
		amount := misc.Ternary[float32](increase, 2, .5)
		switch cursorLine {
		case 0:
			t.kp *= amount
		case 1:
			t.ki *= amount
		case 2:
			t.kd *= amount
		default:
			maxSpd *= amount
		}
		lastErrRate = fmt.Sprintf("%.2f", errAcc/errCount)
		errAcc = 0
		errCount = 0
	}
	updateInput := func() {
		buttonsNew := Buttons()
		if buttonsNew[BUp] && !buttonsOld[BUp] {
			cursorLine = (cursorLine + 3) % 4
		}
		if buttonsNew[BDown] && !buttonsOld[BDown] {
			cursorLine = (cursorLine + 1) % 4
		}
		if buttonsNew[BLeft] && !buttonsOld[BLeft] {
			modifyInput(false)
		}
		if buttonsNew[BRight] && !buttonsOld[BRight] {
			modifyInput(true)
		}
		buttonsOld = buttonsNew
	}
	updateDsp := func() {
		dsp.Clear()
		_, p, _ := misc.Float32Parts(t.kp)
		dsp.Write(0, 0, cursor(0)+" p: "+misc.IToA(int(p)-127))
		_, i, _ := misc.Float32Parts(t.ki)
		dsp.Write(1, 0, cursor(1)+" i: "+misc.IToA(int(i)-127))
		_, d, _ := misc.Float32Parts(t.kd)
		dsp.Write(2, 0, cursor(2)+" d: "+misc.IToA(int(d)-127))
		dsp.Write(3, 0, cursor(3)+" s: "+misc.IToA(int(maxSpd)))
		dsp.Write(4, 0, "e: "+fmt.Sprintf("%.2f", errAcc/errCount))
		dsp.Write(5, 0, "l: "+lastErrRate)
	}
	timeNew := time.Now().UnixMicro()
	timeOld := timeNew - 1
	startPositions := t.positions()
	t.left.Command(MCRunDirect)
	t.right.Command(MCRunDirect)
	timeDsp := timeNew
	for {
		delta := t.positions().Sub(startPositions)
		fasterB := delta[0] < delta[1]
		fasterI := misc.BToI(fasterB)
		excess := delta[fasterI] - delta[1-fasterI]
		errPos := misc.Ternary(fasterB, -excess, excess)
		timeDelta := float32(timeNew - timeOld)
		t.i += errPos * t.ki * timeDelta
		port, correction := misc.SignBitAndMag(t.i + errPos*t.kp + (errPos-t.last)*t.kd/timeDelta)
		dutyCyclesF := misc.Vec[float32]{1, 1}
		dutyCyclesF[port] -= correction
		dutyCyclesI := dutyCyclesF.Mul1(maxSpd / dutyCyclesF.Abs().Max()).RoundI()
		t.left.DutyCycleSetpointSet(dutyCyclesI[0])
		t.right.DutyCycleSetpointSet(dutyCyclesI[1])
		t.last = errPos
		timeOld = timeNew
		timeNew = time.Now().UnixMicro()
		errAcc += misc.Abs(errPos)
		errCount += 1
		updateInput()
		if timeNew > timeDsp {
			log.Println(delta, errPos)
			updateDsp()
			timeDsp += 500000
		}
	}
}
func (t *Tank) Distance(left, right int) {
	distances := misc.Vec[float32]{float32(left), float32(right)}
	// dutyCycleSq := distances. misc.Abs().Max() * t.accK
	dutyCycleSq := distances.Abs().Max() * 50
	timeNew := time.Now().UnixMilli()
	timeOld := timeNew - 1
	startPositions := misc.Ternary(t.modeHold, t.stopPositions, t.positions())
	t.left.Command(MCRunDirect)
	t.right.Command(MCRunDirect)
	for {
		delta := t.positions().Sub(startPositions)
		progress2 := delta.Div(distances)
		fasterB := progress2[0] < progress2[1]
		fasterI := misc.BToI(fasterB)
		progress := progress2[1-fasterI]
		if progress >= 1 {
			break
		}
		excess := delta[fasterI] - distances[fasterI]*progress
		errPos := misc.Ternary(fasterB, -excess, excess)
		timeDelta := float32(timeNew - timeOld)
		t.i += errPos * t.ki * timeDelta
		port, correction := misc.SignBitAndMag(t.i + errPos*t.kp + (errPos-t.last)*t.kd/timeDelta)
		dutyCyclesF := distances.Copy()
		dutyCyclesF[port] -= dutyCyclesF[port] * correction
		// Clamps values magnitudes to maximum while maintaining their proportionality and rounds them.
		// maximum must be positive.
		dutyCyclesI := dutyCyclesF.Mul1(misc.Min(100, misc.Sqrt((1-progress)*dutyCycleSq)) / dutyCyclesF.Abs().Max()).RoundI()
		t.left.DutyCycleSetpointSet(dutyCyclesI[0])
		t.right.DutyCycleSetpointSet(dutyCyclesI[1])
		t.last = errPos
		timeOld = timeNew
		timeNew = time.Now().UnixMilli()
	}
	t.left.Command(MCStop)
	t.right.Command(MCStop)
	t.last = 0
	t.stopPositions = startPositions.Add(distances)
	t.modeHold = true
}
func (t Tank) Straight(mm float32) {
	dst := misc.RoundI(mm * 5)
	t.Distance(dst, dst)
}
