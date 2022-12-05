package main

import (
	"image/png"
	"log"
	"os"
	"syscall"

	misc "github.com/IgneousRed/gomisc"
)

const pixelBytes = 4
const dspWidth = 178
const dspHeight = 128
const dspStride = dspWidth * pixelBytes
const dspBytes = dspHeight * dspStride
const dspChars = 95
const charWidth = 16
const charHeight = 16
const charStride = charWidth * pixelBytes
const charBytes = charHeight * charStride

type Display struct {
	framebuffer []byte
	whiteScreen [dspBytes]byte
	characters  [dspChars][charBytes]byte
}

var displayInitialized bool

func DisplayInit() Display {
	misc.FatalIf(displayInitialized, "Display already initialized")
	var result Display

	// Clear
	for i := 0; i < dspBytes; i += 4 {
		result.whiteScreen[i] = 255
		result.whiteScreen[i+1] = 255
		result.whiteScreen[i+2] = 255
	}

	// Chars
	f, err := os.Open("chars.png")
	misc.FatalErr("", err)
	img, err := png.Decode(f)
	misc.FatalErr("", err)
	f.Close()
	for c := range result.characters {
		xStart := c * (charWidth + 1)
		for y := 0; y < charHeight; y++ {
			cStart := y * charWidth
			for x := 0; x < charWidth; x++ {
				cStart := (cStart + x) * pixelBytes
				r, g, b, _ := img.At(xStart+x, y).RGBA()
				result.characters[c][cStart] = byte(b >> 8)
				result.characters[c][cStart+1] = byte(g >> 8)
				result.characters[c][cStart+2] = byte(r >> 8)
			}
		}
	}

	// Map and Clear
	fb, err := os.OpenFile("/dev/fb0", os.O_RDWR, 0)
	misc.FatalErr("", err)
	result.framebuffer, err = syscall.Mmap(int(fb.Fd()), 0, dspBytes, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	misc.FatalErr("", err)
	copy(result.framebuffer, result.whiteScreen[:])
	displayInitialized = true
	return result
}

// Crears whole screen
func (d Display) Clear() {
	if !displayInitialized {
		log.Fatal("Display not Initialized")
	}
	copy(d.framebuffer, d.whiteScreen[:])
}
func (d Display) Write(posY, posX int, text string) { // TODO: Provide more ways to write
	if !displayInitialized {
		log.Fatal("Display not Initialized")
	}
	misc.FatalIf(posY >= 8, "The display has only 8 lines")
	fbStart := posY * charHeight * dspStride
	for len(text) > 0 && posX < 11 {
		c := text[0] - ' '
		misc.FatalIf(c >= dspChars, "Unprintable character")
		fbStart := fbStart + (posX*charWidth+1)*pixelBytes
		for y := 0; y < charHeight; y++ {
			cStart := y * charStride
			copy(d.framebuffer[fbStart+y*dspStride:], d.characters[c][cStart:cStart+charStride])
		}
		text, posX = text[1:], posX+1
	}
}
