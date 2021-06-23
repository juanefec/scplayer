package component

import (
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/juanefec/gui"
	"github.com/juanefec/gui/win"
	. "github.com/juanefec/scplayer/util"
)

var transparentDarkCyan = color.RGBA{0x00, 0x8b, 0x8b, 0x90}

func Infobar(env gui.Env, theme *Theme, newInfo <-chan string, search func(string)) {

	isOpen := false

	redraw := func(r image.Rectangle, text, icon, info image.Image, isOpen bool) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			if isOpen {
				draw.Draw(drw, r, &image.Uniform{theme.Infobar}, image.Point{}, draw.Src)
				DrawCentered(drw, r, text, draw.Over)
				DrawLeftCentered(drw, r, icon, draw.Over)
				return r
			}
			draw.Draw(drw, r, &image.Uniform{theme.Infobar}, image.Point{}, draw.Src)
			DrawCentered(drw, r, info, draw.Over)
			return r
		}
	}

	var (
		r          image.Rectangle
		searchterm strings.Builder
		icon       image.Image
		text       image.Image

		// info
		info image.Image
	)

	icon = MakeIconImage("search")
	for {
		select {
		case nf := <-newInfo:
			info = MakeTextScaledImage(nf, theme.Face, theme.Text, 0.9)
			env.Draw() <- redraw(r, text, icon, info, isOpen)
		case e, ok := <-env.Events():
			if !ok {
				return
			}
			switch e := e.(type) {
			case gui.Resize:
				r = e.Rectangle
				text = MakeTextImage(searchterm.String(), theme.Face, theme.Text)
				env.Draw() <- redraw(r, text, icon, info, isOpen)

			case win.KbRepeat:
				switch e.Key {
				case win.KeyBackspace:
					if isOpen {
						sofar := searchterm.String()
						if sofar != "" {
							searchterm.Reset()
							searchterm.WriteString(sofar[:len(sofar)-1])
						}

						text = MakeTextImage(searchterm.String(), theme.Face, theme.Text)
						env.Draw() <- redraw(r, text, icon, info, isOpen)
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
					text = MakeTextImage(searchterm.String(), theme.Face, theme.Text)
					env.Draw() <- redraw(r, text, icon, info, isOpen)

				case win.KeyEscape:
					search("")
					searchterm.Reset()
					isOpen = false
					text = MakeTextImage(searchterm.String(), theme.Face, theme.Text)
					env.Draw() <- redraw(r, text, icon, info, isOpen)

				case win.KeyBackspace:
					if isOpen {
						sofar := searchterm.String()
						if sofar != "" {
							searchterm.Reset()
							searchterm.WriteString(sofar[:len(sofar)-1])
						}
					}
					text = MakeTextImage(searchterm.String(), theme.Face, theme.Text)
					env.Draw() <- redraw(r, text, icon, info, isOpen)
				}

			case win.KbType:
				if isOpen && isAlphanumeric(e.Rune) {
					searchterm.WriteRune(e.Rune)
				}
				text = MakeTextImage(searchterm.String(), theme.Face, theme.Text)
				env.Draw() <- redraw(r, text, icon, info, isOpen)
			}

		}
	}

}

const alphanumeric = "qwertyuiopasdfghjklzxcvbnm1234567890-_"

func isAlphanumeric(key rune) bool {
	return strings.ContainsRune(alphanumeric, key)
}
