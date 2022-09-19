package ev3go

import (
	"image"
	"image/draw"

	"github.com/ev3go/ev3"
	"github.com/ev3go/ev3dev"
	"github.com/ev3go/ev3dev/fb"
)

var letter = &fb.Monochrome{
	Pix:    []uint8{},
	Stride: 0,
	Rect: image.Rectangle{
		Min: image.Point{
			X: 0,
			Y: 0,
		},
		Max: image.Point{
			X: 16,
			Y: 16,
		},
	},
}

func LCDInit() {
	ev3.LCD.Init(false)
	_ = ev3dev.Down
}
func LCDClear()
func LCDWrite(posY, posX int, text string) {
	draw.Draw(ev3.LCD, ev3.LCD.Bounds(), letter, letter.Bounds().Min, draw.Src)
}
