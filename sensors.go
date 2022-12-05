package main

import misc "github.com/IgneousRed/gomisc"

const sensorPath = "/sys/class/lego-sensor"

type NoSensor misc.Atom
type Ultrasonic struct {
	path string
}
type Infrared struct {
	path string
}
type Color struct {
	path string
}
type SensorType interface { // remove "Type"?
	NoSensor | Ultrasonic | Infrared | Color
}

var sensorsInitialized bool

// Blocks until requested Sensors are correctly connected.
// Initialize Display before Sensors.
// func SensorsInit[P1, P2, P3, P4 SensorType](dsp Display) (P1, P2, P3, P4) {
// 	FatalIf(sensorsInitialized, "Sensors already Initialized")
// 	ports := [4]SensorType{port1, port2, port3, port4}
// 	for {
// 		// Find connected Sensors
// 		f, err := os.Open(sensorPath)
// 		FatalErr("", err)
// 		names, err := f.Readdirnames(0)
// 		FatalErr("", err)
// 		f.Close()

// 		// Populate conections
// 		connected := [4]SensorType{}
// 		retry := false // If a Sensor disconnects, restart
// 		for _, name := range names {
// 			path := filepath.Join(sensorPath, name, "")
// 			attribute := func(att string) string {
// 				val, err := os.ReadFile(path + att)
// 				if err != nil {
// 					retry = true
// 					return ""
// 				}
// 				return string(val[:len(val)-1]) // Removes "\n" at the end
// 			}
// 			port := attribute("address")
// 			sensor := attribute("driver_name")
// 			if retry {
// 				break
// 			}
// 			log.Println(port, sensor)
// 			var portNum = map[string]int{
// 				"ev3-ports:in1": 0,
// 				"ev3-ports:in2": 1,
// 				"ev3-ports:in3": 2,
// 				"ev3-ports:in4": 3,
// 			}
// 			var sensorType = map[string]SensorType{ // warning
// 				"lego-ev3-us":    Ultrasonic,
// 				"lego-ev3-ir":    Infrared,
// 				"lego-ev3-color": Color,
// 			}
// 			i := portNum[port]
// 			connected[i] = sensorType[sensor]
// 			sensors[i].path = path
// 		}
// 		if retry {
// 			continue
// 		}

// 		// Does mismatch exist
// 		mismatch := false
// 		for i := range sensors {
// 			if connected[i] != ports[i] {
// 				mismatch = true
// 				break
// 			}
// 		}
// 		DisplayClear()
// 		if !mismatch {
// 			break
// 		}

// 		// Prints help
// 		DisplayWrite(1, 2, "Sensors")
// 		for i := range sensors {
// 			var sensorName = [4]string{"None", "Sonic", "Infra", "Color"}
// 			postFix := Ternary(connected[i] != ports[i], sensorName[ports[i]], "Ok")
// 			DisplayWrite(i+2, 0, IToA(i+1)+": "+postFix)
// 		}
// 		time.Sleep(time.Millisecond * 100)
// 	}
// 	sensorsInitialized = true
// }
