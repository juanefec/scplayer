package component

import (
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	. "github.com/juanefec/scplayer/util"
)

var transparentDarkCyan = color.RGBA{0x00, 0x8b, 0x8b, 0x90}

func Searchbar(env gui.Env, theme *Theme, search func(string)) {

	isOpen := false

	redraw := func(r image.Rectangle, text, icon image.Image, isOpen bool) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			if isOpen {
				draw.Draw(drw, r, &image.Uniform{transparentDarkCyan}, image.Point{}, draw.Src)
				DrawCentered(drw, r, text, draw.Over)
				DrawLeftCentered(drw, r, icon, draw.Over)
				return r
			}
			return image.Rect(0, 0, 0, 0)
		}
	}

	var (
		r          image.Rectangle
		searchterm strings.Builder
		icon       image.Image
		text       image.Image
	)

	icon = MakeIconImage("search")

	for e := range env.Events() {
		switch e := e.(type) {
		case gui.Resize:
			r = e.Rectangle
			text = MakeTextImage(searchterm.String(), theme.Face, color.Black)
			env.Draw() <- redraw(r, text, icon, isOpen)

		case win.KbRepeat:
			switch e.Key {
			case win.KeyBackspace:
				if isOpen {
					sofar := searchterm.String()
					if sofar != "" {
						searchterm.Reset()
						searchterm.WriteString(sofar[:len(sofar)-1])
					}

					text = MakeTextImage(searchterm.String(), theme.Face, color.Black)
					env.Draw() <- redraw(r, text, icon, isOpen)
				}
			}

		case win.KbUp:
			switch e.Key {
			case win.KeyEnter:
				if isOpen {
					search(searchterm.String())
					searchterm.Reset()
				}
				isOpen = !isOpen
				text = MakeTextImage(searchterm.String(), theme.Face, color.Black)
				env.Draw() <- redraw(r, text, icon, isOpen)

			case win.KeyEscape:
				search("")
				searchterm.Reset()
				isOpen = false
				text = MakeTextImage(searchterm.String(), theme.Face, color.Black)
				env.Draw() <- redraw(r, text, icon, isOpen)

			case win.KeyBackspace:
				if isOpen {
					sofar := searchterm.String()
					if sofar != "" {
						searchterm.Reset()
						searchterm.WriteString(sofar[:len(sofar)-1])
					}
				}
				text = MakeTextImage(searchterm.String(), theme.Face, color.Black)
				env.Draw() <- redraw(r, text, icon, isOpen)
			}

		case win.KbType:
			if isOpen && isAlphanumeric(e.Rune) {
				searchterm.WriteRune(e.Rune)
			}
			text = MakeTextImage(searchterm.String(), theme.Face, color.Black)
			env.Draw() <- redraw(r, text, icon, isOpen)
		default:
			env.Draw() <- redraw(r, text, icon, isOpen)
		}
	}
	close(env.Draw())
}

const alphanumeric = "qwertyuiopasdfghjklzxcvbnm1234567890-"

func isAlphanumeric(key rune) bool {
	return strings.ContainsRune(alphanumeric, key)
}
