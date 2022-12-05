package main

import (
	"github.com/ev3go/ev3dev"
)

//	func getEverything() {
//		log.Println("get")
//		log.Println(DutyCycleSetpointGet(MPB))
//		log.Println(HoldPidGet(MPB, PKD))
//		log.Println(HoldPidGet(MPB, PKI))
//		log.Println(HoldPidGet(MPB, PKP))
//		log.Println(PolarityGet(MPB))
//		log.Println(PositionGet(MPB))
//		log.Println(PositionSetpointGet(MPB))
//		log.Println(RampDownGet(MPB))
//		log.Println(RampUpGet(MPB))
//		log.Println(SpeedPidGet(MPB, PKD))
//		log.Println(SpeedPidGet(MPB, PKI))
//		log.Println(SpeedPidGet(MPB, PKP))
//		log.Println(SpeedSetpointGet(MPB))
//		log.Println(StopActionGet(MPB))
//		log.Println(TimeSetpointGet(MPB))
//	}
//
//	func setEverything() {
//		log.Println("set")
//		DutyCycleSetpointSet(MPB, 69)
//		HoldPidSet(MPB, PKD, 69)
//		HoldPidSet(MPB, PKI, 69)
//		HoldPidSet(MPB, PKP, 69)
//		PolaritySet(MPB, PInversed) // 1
//		PositionSet(MPB, 69)
//		PositionSetpointSet(MPB, 69)
//		RampDownSet(MPB, 69)
//		RampUpSet(MPB, 69)
//		SpeedPidSet(MPB, PKD, 69)
//		SpeedPidSet(MPB, PKI, 69)
//		SpeedPidSet(MPB, PKP, 69)
//		SpeedSetpointSet(MPB, 69)
//		StopActionSet(MPB, SAHold) // 2
//		TimeSetpointSet(MPB, 69)
//	}
//
//	func testSpeed() {
//		MotorCommand(MPB, CRunDirect)
//		MotorCommand(MPC, CRunDirect)
//		for p := 10; p <= 100; p += 10 {
//			DutyCycleSetpointSet(MPB, p)
//			DutyCycleSetpointSet(MPC, p)
//			time.Sleep(time.Second * 2)
//			log.Println(p, " -> ", Speed(MPB), Speed(MPC))
//		}
//		MotorCommand(MPB, CStop)
//		MotorCommand(MPC, CStop)
//	}

// TODO: save settings

func main() {
	_ = ev3dev.Back
	dsp := DisplayInit()
	_, mB, mC, _ := MotorsInit(MTNone, MTLarge, MTLarge, MTNone, dsp)
	tank := TankInit(mB, mC)
	// tank.TuneDeceleration(dsp)
	tank.Distance(360*3, 360*3)
	// time.Sleep(time.Second * 999)
}
