package component

import (
	"image"
	"image/draw"
	"math"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	. "github.com/juanefec/scplayer/util"
)

func VolumeSlider(env gui.Env, theme *Theme, trigUpdateVolume <-chan struct{}, changeVolume func(float64)) {
	volLevels := []image.Image{
		MakeIconImage("volume_lvl-0"),
		MakeIconImage("volume_lvl-1"),
		MakeIconImage("volume_lvl-2"),
		MakeIconImage("volume_lvl-3"),
		MakeIconImage("volume_lvl-4"),
		MakeIconImage("volume_lvl-5"),
		MakeIconImage("volume_lvl-6"),
		MakeIconImage("volume_lvl-7"),
		MakeIconImage("volume_lvl-8"),
		MakeIconImage("volume_lvl-9"),
	}
	vol := 4
	volf := float64(vol)
	volmin, volmax := 0, 9

	redraw := func(r image.Rectangle, over, pressed bool) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			col := theme.VolumeBg
			if over {
				col = theme.VolumeBgOver
			}

			draw.Draw(drw, r, &image.Uniform{col}, image.Point{}, draw.Src)
			DrawCentered(drw, r, volLevels[vol], draw.Over)
			return r
		}
	}

	var (
		r       image.Rectangle
		over    bool
		pressed bool
	)

	for {
		select {
		case <-trigUpdateVolume:
			changeVolume(volf)
		case e := <-env.Events():
			switch e := e.(type) {
			case win.MoMove:
				over = e.Point.In(r)
				if over && pressed {
					volf = MapIntFloat(e.Point.Y, r.Min.Y, r.Max.Y, volmax, volmin)
					if volf > 0.3 {
						vol = int(math.Ceil(volf))
					} else {
						vol = 0
					}
					changeVolume(volf)
				}
				env.Draw() <- redraw(r, over, pressed)
			case gui.Resize:
				r = e.Rectangle
				env.Draw() <- redraw(r, over, pressed)

			case win.MoDown:
				newPressed := e.Point.In(r)
				if newPressed != pressed {
					pressed = newPressed
					volf = MapIntFloat(e.Point.Y, r.Min.Y, r.Max.Y, volmax, volmin)
					if volf > 0.3 {
						vol = int(math.Ceil(volf))
					} else {
						vol = 0
					}
					changeVolume(volf)
					env.Draw() <- redraw(r, over, pressed)
				}

			case win.MoUp:
				if pressed {
					if e.Point.In(r) {
						changeVolume(volf)
					}
					pressed = false
					env.Draw() <- redraw(r, over, pressed)
				}
			}
		}

	}

	//close(env.Draw())
}
