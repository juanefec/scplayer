package component

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/juanefec/gui"
	"github.com/juanefec/gui/win"
	. "github.com/juanefec/scplayer/util"
)

func SelectGeneric(mux *gui.Mux, minI, maxI, n, minY, maxY int, theme *Theme, action func(string), startOption string, options ...string) {

	var (
		view            = make(chan string)
		selected string = startOption
	)

	total := len(options)
	i := 0
	optionMap := map[string]chan struct{}{}
	for _, op := range options {
		ch := make(chan struct{})

		startSelected := false
		if op == startOption {
			startSelected = true
		}

		go SelectButton(EvenVerticalMinMaxY(EvenHorizontal(mux.MakeEnv(), i, i+1, total), minI, maxI, n, minY, maxY), theme, op, ch, startSelected, func(opName string) {
			view <- opName
		})

		optionMap[op] = ch
		i++
	}

	for ns := range view {
		if ns != selected {
			action(ns)
			optionMap[selected] <- struct{}{}
			selected = ns
		}
	}
}

func SelectButton(env gui.Env, theme *Theme, name string, unselect chan struct{}, startSelected bool, action func(string)) {
	iconImg := MakeIconImage(name)

	redraw := func(r image.Rectangle, over, pressed, selected bool) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			var clr color.Color
			if pressed {
				clr = theme.TextBoxUp
			} else if selected {
				clr = theme.Background
			} else if over {
				clr = theme.TextBoxOver
			} else {
				clr = theme.Title
			}
			draw.Draw(drw, r, &image.Uniform{clr}, image.Point{}, draw.Src)
			DrawCentered(drw, r, iconImg, draw.Over)
			return r
		}
	}

	var (
		r        image.Rectangle
		over     bool
		pressed  bool
		selected bool = startSelected
	)

	for {
		select {
		case <-unselect:
			selected = false
			env.Draw() <- redraw(r, over, pressed, selected)
		case e := <-env.Events():
			switch e := e.(type) {
			case win.MoMove:
				over = e.Point.In(r)
				env.Draw() <- redraw(r, over, pressed, selected)
			case gui.Resize:
				r = e.Rectangle
				env.Draw() <- redraw(r, over, pressed, selected)

			case win.MoDown:
				newPressed := e.Point.In(r)
				if newPressed != pressed {
					pressed = newPressed
					env.Draw() <- redraw(r, over, pressed, selected)
				}

			case win.MoUp:
				if pressed {
					if e.Point.In(r) {
						action(name)
						selected = true
					}
					pressed = false
					env.Draw() <- redraw(r, over, pressed, selected)
				}
			}
		}
	}
}
