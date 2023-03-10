package main

import (
	"image/png"
	"os"
	"syscall"

	m "github.com/IgneousRed/gomisc"
)

const pixelBytes = 4 // RGBA
const dspWidth = 178
const dspHeight = 128
const dspStride = dspWidth * pixelBytes
const dspBytes = dspHeight * dspStride
const dspChars = 95 // Printable characters
const charWidth = 16
const charHeight = 16
const charStride = charWidth * pixelBytes
const charBytes = charHeight * charStride

var displayInitialized bool

// Create using `DisplayInit`
type Display struct {
	framebuffer []byte
	whiteScreen [dspBytes]byte
	characters  [dspChars][charBytes]byte
}

func DisplayInit() Display {
	m.FatalIf(displayInitialized, "Display already initialized")
	var result Display

	// Clear
	for i := 0; i < dspBytes; i += pixelBytes {
		result.whiteScreen[i] = 255
		result.whiteScreen[i+1] = 255
		result.whiteScreen[i+2] = 255
	}

	// Chars
	f, err := os.Open("chars.png")
	m.FatalErr("", err)
	img, err := png.Decode(f)
	m.FatalErr("", err)
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
	m.FatalErr("", err)
	result.framebuffer, err = syscall.Mmap(int(fb.Fd()), 0, dspBytes,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED,
	)
	m.FatalErr("", err)
	result.Clear()
	displayInitialized = true
	return result
}

// Crears whole screen
func (d Display) Clear() {
	copy(d.framebuffer, d.whiteScreen[:])
}

// Writes `text` on the display
func (d Display) Write(posX, posY int, text string) {
	m.FatalIf(posY >= 8, "The display has only 8 lines")
	fbStart := posY * charHeight * dspStride
	for len(text) > 0 && posX < 11 {
		c, fbStart := text[0]-' ', fbStart+(posX*charWidth+1)*pixelBytes
		m.FatalIf(c >= dspChars, "Unprintable character")
		for y := 0; y < charHeight; y++ {
			cStart := y * charStride
			copy(d.framebuffer[fbStart+y*dspStride:],
				d.characters[c][cStart:cStart+charStride],
			)
		}
		text, posX = text[1:], posX+1
	}
}
