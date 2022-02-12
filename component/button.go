package component

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	. "github.com/juanefec/scplayer/util"
)

func Button(env gui.Env, theme *Theme, icon string, action func()) {
	iconImg := MakeIconImage(icon)

	redraw := func(r image.Rectangle, over, pressed bool) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			var clr color.Color
			if pressed {
				clr = theme.ButtonDown
			} else if over {
				clr = theme.ButtonOver
			} else {
				clr = theme.ButtonUp
			}
			draw.Draw(drw, r, &image.Uniform{clr}, image.Point{}, draw.Src)
			DrawCentered(drw, r, iconImg, draw.Over)
			return r
		}
	}

	var (
		r       image.Rectangle
		over    bool
		pressed bool
	)

	for e := range env.Events() {
		switch e := e.(type) {
		case win.MoMove:
			over = e.Point.In(r)
			env.Draw() <- redraw(r, over, pressed)
		case gui.Resize:
			r = e.Rectangle
			env.Draw() <- redraw(r, over, pressed)

		case win.MoDown:
			newPressed := e.Point.In(r)
			if newPressed != pressed {
				pressed = newPressed
				env.Draw() <- redraw(r, over, pressed)
			}

		case win.MoUp:
			if pressed {
				if e.Point.In(r) {
					action()
				}
				pressed = false
				env.Draw() <- redraw(r, over, pressed)
			}
		}
	}

	close(env.Draw())
}

func PauseButton(env gui.Env, theme *Theme, pausebtn <-chan bool, action func(bool)) {
	iconPlayImg := MakeIconImage("play")
	iconPauseImg := MakeIconImage("pause")

	playing := false
	iconImg := iconPlayImg

	redraw := func(r image.Rectangle, over, pressed bool) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			var clr color.Color
			if pressed {
				clr = theme.ButtonDown
			} else if over {
				clr = theme.ButtonOver
			} else {
				clr = theme.ButtonUp
			}
			if playing {
				iconImg = iconPauseImg
			} else {
				iconImg = iconPlayImg
			}
			draw.Draw(drw, r, &image.Uniform{clr}, image.Point{}, draw.Src)
			DrawCentered(drw, r, iconImg, draw.Over)
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
		case p := <-pausebtn:
			playing = p
			env.Draw() <- redraw(r, over, pressed)
		case e, ok := <-env.Events():
			if !ok {
				close(env.Draw())
				return
			}

			switch e := e.(type) {
			case win.MoMove:
				over = e.Point.In(r)
				env.Draw() <- redraw(r, over, pressed)
			case gui.Resize:
				r = e.Rectangle
				env.Draw() <- redraw(r, over, pressed)

			case win.MoDown:
				newPressed := e.Point.In(r)
				if newPressed != pressed {
					pressed = newPressed
					env.Draw() <- redraw(r, over, pressed)
				}

			case win.MoUp:
				if pressed {
					if e.Point.In(r) {
						action(playing)
						playing = !playing
					}
					pressed = false
					env.Draw() <- redraw(r, over, pressed)
				}
			case win.KbDown:
				switch e.Key {
				case win.KeySpace:

					pressed = true
					env.Draw() <- redraw(r, over, pressed)

				}

			case win.KbUp:
				switch e.Key {
				case win.KeySpace:
					if pressed {
						action(playing)
						playing = !playing

						pressed = false

						env.Draw() <- redraw(r, over, pressed)
					}
				}
			}

		}
	}
}
