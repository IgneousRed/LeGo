package main

import (
	"io"
	"os"
	"sync"

	m "github.com/IgneousRed/gomisc"
)

type Button u8

const (
	BRight Button = iota
	BUp
	BLeft
	BDown
	BMiddle
)

var buttons [5]bool
var buttonsMutex sync.RWMutex

func init() {
	file, err := os.Open("/dev/input/by-path/platform-gpio_keys-event")
	m.FatalErr("", err)
	go func() {
		defer file.Close()
		buf, keyToButton := [16]u8{}, [109]u8{
			28:  u8(BMiddle),
			106: u8(BRight),
			103: u8(BUp),
			105: u8(BLeft),
			108: u8(BDown),
		}
		for {
			_, err := io.ReadFull(file, buf[:])
			m.FatalErr("", err)
			if key := buf[10]; key != 0 {
				buttonsMutex.Lock()
				buttons[keyToButton[key]] = m.NToB(buf[12])
				buttonsMutex.Unlock()
			}
		}
	}()
}
func Buttons() [5]bool {
	buttonsMutex.RLock()
	result := buttons
	buttonsMutex.RUnlock()
	return result
}
