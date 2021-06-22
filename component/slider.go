package component

import (
	"image"
	"image/color"
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

func BrowserSlider(env gui.Env, theme *Theme, listenBrowser <-chan int, updateBrowser chan<- int, listenNewBrowser <-chan int, playingPos <-chan int) {
	var highr image.Rectangle
	redraw := func(r image.Rectangle, over, pressed bool, browserPos image.Point, browserMaxY image.Point, playPos int) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			col := theme.VolumeBg
			if over {
				col = theme.VolumeBgOver
			}

			draw.Draw(drw, r, &image.Uniform{col}, image.Point{}, draw.Src)
			mpos := Map(browserPos.Y, 0, browserMaxY.Y, r.Min.Y, r.Max.Y)
			mposY := Map(r.Dy(), 0, browserMaxY.Y, r.Min.Y, r.Max.Y)
			highr = image.Rect(
				r.Min.X,
				mpos,
				r.Max.X,
				mpos+mposY-r.Min.Y,
			)
			draw.DrawMask(
				drw, highr.Intersect(r),
				&image.Uniform{theme.HighlightSlider}, image.Point{},
				&image.Uniform{color.Alpha{64}}, image.Point{},
				draw.Over,
			)
			playpos := Map(playPos, 0, browserMaxY.Y, r.Min.Y, r.Max.Y)
			currsong := image.Rect(
				r.Min.X,
				playpos,
				r.Max.X,
				playpos+4,
			)
			draw.DrawMask(
				drw, currsong.Intersect(r),
				&image.Uniform{theme.HighlightSlider}, image.Point{},
				&image.Uniform{color.Alpha{94}}, image.Point{},
				draw.Over,
			)

			return r
		}
	}

	var (
		r           image.Rectangle
		rclickable  image.Rectangle
		over        bool
		pressed     bool
		browserPos  image.Point
		browserMaxY image.Point

		playPos int

		//pressedAt int
	)

	move := func(p image.Point) int {
		m := Map(p.Y, r.Min.Y, r.Max.Y, 0, browserMaxY.Y)
		return m
	}

	for {
		select {
		case y := <-listenNewBrowser:
			browserMaxY = image.Point{0, y}
			browserPos = image.Point{0, r.Min.Y}
			env.Draw() <- redraw(r, over, pressed, browserPos, browserMaxY, playPos)
			rclickable = r
			rclickable.Max.Y = r.Max.Y - highr.Dy()
		case pp := <-playingPos:
			playPos = pp
			env.Draw() <- redraw(r, over, pressed, browserPos, browserMaxY, playPos)
		case y := <-listenBrowser:
			browserPos = image.Point{0, y}
			env.Draw() <- redraw(r, over, pressed, browserPos, browserMaxY, playPos)
		case e := <-env.Events():
			switch e := e.(type) {
			case win.MoMove:
				rclickable = r
				rclickable.Max.Y = r.Max.Y - highr.Dy()
				over = e.Point.In(rclickable)
				if over && pressed {

					mto := move(e.Point)
					browserPos.Y = mto
					updateBrowser <- mto

				}
				env.Draw() <- redraw(r, over, pressed, browserPos, browserMaxY, playPos)

			case win.MoDown:
				rclickable = r
				rclickable.Max.Y = r.Max.Y - highr.Dy()
				newPressed := e.Point.In(rclickable)
				if newPressed != pressed {
					pressed = newPressed
					mto := move(e.Point)
					browserPos.Y = mto
					//pressedAt = e.Point.Y
					env.Draw() <- redraw(r, over, pressed, browserPos, browserMaxY, playPos)
					updateBrowser <- mto
				}

			case win.MoUp:
				if pressed {
					pressed = false
					env.Draw() <- redraw(r, over, pressed, browserPos, browserMaxY, playPos)
				}

			case gui.Resize:
				r = e.Rectangle
				rclickable = r
				rclickable.Max.Y = r.Max.Y - highr.Dy()
				env.Draw() <- redraw(r, over, pressed, browserPos, browserMaxY, playPos)
			}
		}

	}

	//close(env.Draw())
}
