package main

import (
	"fmt"
	"log"
	"os"
	"time"

	m "github.com/IgneousRed/gomisc"
)

const accPath = "/home/robot/ev3go/tank-acc"
const spdPath = "/home/robot/ev3go/tank-spd"

var tankInitialized bool

type Tank struct {
	motors                     [2]Motor
	kp, ki, kd                 f64
	mmToMDeg                   f64
	k_mdegPrSMicroS, k_mdegPrS f64
	mdegPrSMicroS, mdegPrS     f64
	modeDist                   bool
	start_mdeg                 v2
	i, last                    f64
}

func (t Tank) StopActionSet(action StopAction) {
	t.motors[0].StopActionSet(action)
	t.motors[1].StopActionSet(action)
}
func (t Tank) Command(command MotorCommand) {
	t.motors[0].Command(command)
	t.motors[1].Command(command)
}
func readDefault(path string, defaultValue f64) f64 {
	data, err := os.ReadFile(path)
	if err != nil {
		m.FatalErr("", os.WriteFile(path, m.F64ToU8s(defaultValue), os.ModePerm))
		return defaultValue
	}
	return m.U8sToF64(data)
}
func TankInit(left, right Motor) Tank {
	m.FatalIf(tankInitialized, "Tank already Initialized")
	var result Tank
	result.motors = [2]Motor{left, right}
	result.kp = m.Pow(2., -8.)
	result.ki = m.Pow(2., -30.)
	result.kd = m.Pow(2., -1.)
	result.k_mdegPrSMicroS = readDefault(accPath, 2)
	result.k_mdegPrS = readDefault(spdPath, 700)
	result.mdegPrSMicroS = result.k_mdegPrSMicroS * .5
	result.mdegPrS = result.k_mdegPrS * .8
	// result.modeDist = true
	result.StopActionSet(SAHold)
	result.Command(MCStop)
	return result
}
func (t Tank) Positions() (mdeg v2) {
	return V2(f64(t.motors[0].PositionGet()), f64(t.motors[1].PositionGet()))
}
func (t Tank) SpeedGet() (mdegPrS v2) {
	return V2(f64(t.motors[0].Speed()), f64(t.motors[1].Speed()))
}
func (t *Tank) SpeedSet(mdegPrS v2) {
	t.motors[0].SpeedTargetSet(int(mdegPrS[0]))
	t.motors[1].SpeedTargetSet(int(mdegPrS[1]))
	t.Command(MCRunForever)
	t.modeDist = false
}
func (t *Tank) Stop() {
	if t.motors[0].State() == SHolding {
		log.Println("Warning: Tank already stopped")
		return
	}
	t.Command(MCStop)
}
func (t Tank) waitSpeedStabilize() {
	old_mdegPrS, dir, count := 0., 1., 5
	for {
		new_mdegPrS := t.SpeedGet().Sum()
		if (new_mdegPrS-old_mdegPrS)*dir < 0 {
			if count < 1 {
				break
			}
			dir, count = dir*-1, count-1
		}
		old_mdegPrS = new_mdegPrS
	}
}

// Blocks
func (t Tank) TuneAccAndSpd(dsp Display) {
	t.motors[0].PowerTargetSet(100)
	t.motors[1].PowerTargetSet(100)
	t.Command(MCRunDirect)
	t.waitSpeedStabilize()
	mdegPrS := t.SpeedGet().Min()
	t.SpeedSet(V2(0, 0))
	for t.SpeedGet().Max() > 0 {
	}
	time.Sleep(time.Millisecond * 100)
	start_microS := MicroS()
	t.SpeedSet(V2(mdegPrS, mdegPrS))
	t.waitSpeedStabilize()
	mdegPrSMicroS := mdegPrS / f64(MicroS()-start_microS)
	t.SpeedSet(V2(0, 0))
	// TODO: Errors
	os.WriteFile(accPath, m.F64ToU8s(mdegPrSMicroS), os.ModePerm)
	os.WriteFile(spdPath, m.F64ToU8s(mdegPrS), os.ModePerm)
	dsp.Write(0, 0, "New Acc:")
	dsp.Write(0, 1, fmt.Sprint(mdegPrSMicroS))
	dsp.Write(0, 2, "New Spd:")
	dsp.Write(0, 3, fmt.Sprint(mdegPrS))
	time.Sleep(time.Second * 1000000)
}
func (t *Tank) TunePID(dsp Display) {
	// maxSpd := f64(100)
	// cursorLine := 0
	// errAcc := f64(0)
	// errCount := f64(0)
	// lastErrRate := ""
	// buttonsOld := Buttons()
	// cursor := func(y int) string {
	// 	if y == cursorLine {
	// 		return ">"
	// 	}
	// 	return " "
	// }
	// modifyInput := func(increase bool) {
	// 	amount := m.Ternary[f64](increase, 2, .5)
	// 	switch cursorLine {
	// 	case 0:
	// 		t.kp *= amount
	// 	case 1:
	// 		t.ki *= amount
	// 	case 2:
	// 		t.kd *= amount
	// 	default:
	// 		maxSpd *= amount
	// 	}
	// 	lastErrRate = fmt.Sprintf("%.2f", errAcc/errCount)
	// 	errAcc = 0
	// 	errCount = 0
	// }
	// updateInput := func() {
	// 	buttonsNew := Buttons()
	// 	if buttonsNew[BUp] && !buttonsOld[BUp] {
	// 		cursorLine = (cursorLine + 3) % 4
	// 	}
	// 	if buttonsNew[BDown] && !buttonsOld[BDown] {
	// 		cursorLine = (cursorLine + 1) % 4
	// 	}
	// 	if buttonsNew[BLeft] && !buttonsOld[BLeft] {
	// 		modifyInput(false)
	// 	}
	// 	if buttonsNew[BRight] && !buttonsOld[BRight] {
	// 		modifyInput(true)
	// 	}
	// 	buttonsOld = buttonsNew
	// }
	// updateDsp := func() {
	// 	dsp.Clear()
	// 	_, p, _ := m.F32ToParts(t.kp)
	// 	dsp.Write(0, 0, cursor(0)+" p: "+m.IToA(int(p)-127))
	// 	_, i, _ := m.F32ToParts(t.ki)
	// 	dsp.Write(1, 0, cursor(1)+" i: "+m.IToA(int(i)-127))
	// 	_, d, _ := m.F32ToParts(t.kd)
	// 	dsp.Write(2, 0, cursor(2)+" d: "+m.IToA(int(d)-127))
	// 	dsp.Write(3, 0, cursor(3)+" s: "+m.IToA(int(maxSpd)))
	// 	dsp.Write(4, 0, "e: "+fmt.Sprintf("%.2f", errAcc/errCount))
	// 	dsp.Write(5, 0, "l: "+lastErrRate)
	// }
	// timeNew := Micros()
	// timeOld := timeNew - 1
	// startPositions := t.positions()
	// t.left.Command(MCRunDirect)
	// t.right.Command(MCRunDirect)
	// timeDsp := timeNew
	// for {
	// 	delta := t.positions().Sub(startPositions)
	// 	fasterB := delta[0] < delta[1]
	// 	fasterI := m.BToI(fasterB)
	// 	excess := delta[fasterI] - delta[1-fasterI]
	// 	errPos := m.Ternary(fasterB, -excess, excess)
	// 	timeDelta := f64(timeNew - timeOld)
	// 	t.i += errPos * t.ki * timeDelta
	// 	port, correction := m.SignBitAndMag(t.i + errPos*t.kp + (errPos-t.last)*t.kd/timeDelta)
	// 	dutyCyclesF := v2{1, 1}
	// 	dutyCyclesF[port] -= correction
	// 	dutyCyclesI := dutyCyclesF.Mul1(maxSpd / dutyCyclesF.Abs().Max()).RoundI()
	// 	t.left.DutyCycleTargetSet(dutyCyclesI[0])
	// 	t.right.DutyCycleTargetSet(dutyCyclesI[1])
	// 	t.last = errPos
	// 	timeOld = timeNew
	// 	timeNew = Micros()
	// 	errAcc += m.Abs(errPos)
	// 	errCount += 1
	// 	updateInput()
	// 	if timeNew > timeDsp {
	// 		log.Println(delta, errPos)
	// 		updateDsp()
	// 		timeDsp += 500000
	// 	}
	// }
}
func (t *Tank) Distance(distances_mdeg v2) {
	start_mdeg := m.Ternary(t.modeDist, t.start_mdeg, t.Positions())
	defer func() {
		t.Command(MCStop)
		t.modeDist = true
		t.start_mdeg = start_mdeg.Add(distances_mdeg)
		t.last = 0
	}()
	absSum_mdeg := distances_mdeg.Abs().Sum()
	if absSum_mdeg == 0 {
		return
	}
	distancesSign := distances_mdeg.Sign()
	mdegSqPrSSq := distances_mdeg.Abs().Max() * t.mdegPrSMicroS * 2 * 1_000_000
	start_microS := MicroS()
	new_microS, old_microS := start_microS, start_microS-1
	for {
		delta_mdeg := t.Positions().Sub(start_mdeg)
		distancesLeft := distances_mdeg.Sub(delta_mdeg)
		progressLeft := distancesLeft.Abs().Sum() / absSum_mdeg
		if distancesLeft.Mul(distancesSign).Sum() <= 0 {
			break
		}
		err_mdeg := delta_mdeg[0] - distances_mdeg[0]*(1-progressLeft)
		delta_microS := f64(new_microS - old_microS)
		t.i += err_mdeg * t.ki * delta_microS
		port, correction := m.SignBitAndMag(t.i + err_mdeg*t.kp +
			(err_mdeg-t.last)*t.kd/delta_microS,
		)
		speeds := distances_mdeg
		speeds[port] -= speeds[port] * correction
		t.SpeedSet(speeds.Mul1(m.Min(
			f64(new_microS-start_microS)*t.mdegPrSMicroS,
			t.mdegPrS,
			m.Sqrt(progressLeft*mdegSqPrSSq),
		) / speeds.Abs().Max())) // Div 0?
		t.last, new_microS, old_microS = err_mdeg, MicroS(), new_microS
	}
}

func (t Tank) Straight(mm f64) {
	mdeg := mm * t.mmToMDeg
	t.Distance(V2(mdeg, mdeg))
}

func (t Tank) Turn(deg f64) {
	// t.Distance()
}

// _angle_deg = 90
// Sub Turn ' Initialize()
//   _distances_mtrDeg[0] = _angle_deg * WHL_DST_HLF_CM * MTRDEG_PR_CM_DEG
//   _distances_mtrDeg[1] = -_distances_mtrDeg[0]
//   Move()
// EndSub
