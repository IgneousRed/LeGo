package main

import (
	"encoding/binary"
	"io"
	"os"
	"sync"

	misc "github.com/IgneousRed/gomisc"
)

type Button byte

const (
	BUp = Button(iota)
	BLeft
	BMiddle
	BRight
	BDown
)

var buttons [5]bool
var buttonsMutex sync.RWMutex

func init() {
	file, err := os.Open("/dev/input/by-path/platform-gpio_keys-event")
	misc.FatalErr("", err)
	go func() {
		var keyToButton = [109]byte{
			103: 0,
			105: 1,
			28:  2,
			106: 3,
			108: 4,
		}
		var buf [16]byte
		for {
			_, err := io.ReadFull(file, buf[:])
			misc.FatalErr("", err)
			key, _ := binary.Uvarint(buf[10:12])
			if key != 0 {
				val, _ := binary.Uvarint(buf[12:16])
				buttonsMutex.Lock()
				buttons[keyToButton[key]] = misc.IToB(val)
				buttonsMutex.Unlock()
			}
		}
	}()
}
func Buttons() [5]bool {
	buttonsMutex.RLock()
	temp := buttons
	buttonsMutex.RUnlock()
	return temp
}
