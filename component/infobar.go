package component

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"
	"time"

	"github.com/juanefec/gui"
	"github.com/juanefec/gui/win"
	"github.com/juanefec/scplayer/sc"
	. "github.com/juanefec/scplayer/util"
)

var transparentDarkCyan = color.RGBA{0x00, 0x8b, 0x8b, 0x90}

func Infobar(env gui.Env, theme *Theme, newInfo <-chan string, newListeningTime <-chan string, search func(string)) {

	redraw := func(r image.Rectangle, listeningTime, info image.Image) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, &image.Uniform{theme.Infobar}, image.Point{}, draw.Src)
			textRect := r
			textRect.Min.X = textRect.Min.X + 10 // :)
			DrawLeftCentered(drw, textRect, info, draw.Over)
			timeRect := r
			timeRect.Max.X = timeRect.Max.X - 10 // :)
			DrawRightCentered(drw, timeRect, listeningTime, draw.Over)
			return r
		}
	}

	var (
		r             image.Rectangle
		info          image.Image
		listeningTime image.Image
	)

	for {
		select {
		case lt := <-newListeningTime:
			listeningTime = MakeTextImage(fmt.Sprintf("spent: %v", lt), theme.Face, theme.Text)
			env.Draw() <- redraw(r, listeningTime, info)
		case nf := <-newInfo:
			info = MakeTextImage(nf, theme.Face, theme.Text)
			env.Draw() <- redraw(r, listeningTime, info)
		case e, ok := <-env.Events():
			if !ok {
				return
			}
			switch e := e.(type) {
			case gui.Resize:
				r = e.Rectangle
				env.Draw() <- redraw(r, listeningTime, info)
			}

		}
	}

}

func Searchbar(env gui.Env, theme *Theme, search func(string)) {

	redraw := func(r image.Rectangle, over, pressed bool, text, avatar, icon, info, cursor, phtext image.Image, isOpen, showCursor bool) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {

			var clr color.Color
			if pressed {
				clr = theme.TextBoxDown
			} else if over {
				clr = theme.TextBoxOver
			} else {
				clr = theme.TextBoxUp
			}
			draw.Draw(drw, r, &image.Uniform{clr}, image.Point{}, draw.Src)

			if isOpen {
				textRect := r
				textRect.Min.X = textRect.Min.X + icon.Bounds().Dx() + 5 // :)
				DrawLeftCentered(drw, textRect, text, draw.Over)
				DrawLeftCentered(drw, r, icon, draw.Over)
				if showCursor {
					nr := r
					nr.Min.X += text.Bounds().Max.X
					draw.Draw(drw, nr, cursor, r.Min, draw.Over)
				}
			} else {
				DrawLeftCentered(drw, r, avatar, draw.Over)
				DrawCentered(drw, r, phtext, draw.Over)
			}
			return r
		}
	}

	var (
		r                                 image.Rectangle
		emptyImg                          = image.NewRGBA(r)
		searchterm                        strings.Builder
		icon                              image.Image
		text                              image.Image
		phtext                            image.Image
		cursor                            image.Image
		avatar                            image.Image = emptyImg
		over, pressed, isOpen, showCursor bool
		exitCursor                        = make(chan struct{})
		sentSearch                        string
		err                               error

		// info
		info image.Image
	)
	cursor = MakeCursorImage(r, color.White)
	animateCursor := func(exit <-chan struct{}) {
		intervalc := 0
		on := true
		for {
			select {
			case <-exit:
				return
			case <-time.After(time.Millisecond * 10):
				if intervalc > 30 {
					intervalc = 0
					if on {
						showCursor = false
					} else {
						showCursor = true
					}
					on = !on
				}

				env.Draw() <- redraw(r, over, pressed, text, avatar, icon, info, cursor, phtext, isOpen, showCursor)

				intervalc++
			}
		}
	}

	getAvatar := func(user string) {
		avatar = emptyImg
		avatar, err = sc.GetAvatar(user)
		if err != nil {
			fmt.Println(err)
		}
	}

	icon = MakeIconImage("search")

	for {
		select {
		case e, ok := <-env.Events():
			if !ok {
				return
			}
			switch e := e.(type) {
			case gui.Resize:
				r = e.Rectangle
				cursor = MakeCursorImage(r, color.White)
				text = MakeTextImage(searchterm.String(), theme.Face, theme.Text)
				env.Draw() <- redraw(r, over, pressed, text, avatar, icon, info, cursor, phtext, isOpen, showCursor)

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
						env.Draw() <- redraw(r, over, pressed, text, avatar, icon, info, cursor, phtext, isOpen, showCursor)
					}
				}

			case win.KbUp:
				switch e.Key {
				case win.KeyEnter:
					st := searchterm.String()
					if isOpen {
						search(st)
						if st != "" {
							sentSearch = st
							go getAvatar(sentSearch)
						}
						searchterm.Reset()
						exitCursor <- struct{}{}
					} else {
						go animateCursor(exitCursor)
					}
					isOpen = !isOpen
					phtext = MakeTextImage(sentSearch, theme.Face, theme.Text)
					text = MakeTextImage("", theme.Face, theme.Text)
					env.Draw() <- redraw(r, over, pressed, text, avatar, icon, info, cursor, phtext, isOpen, showCursor)

				case win.KeyEscape:
					search("")
					searchterm.Reset()
					isOpen = false
					text = MakeTextImage(searchterm.String(), theme.Face, theme.Text)
					env.Draw() <- redraw(r, over, pressed, text, avatar, icon, info, cursor, phtext, isOpen, showCursor)

				case win.KeyBackspace:
					if isOpen {
						sofar := searchterm.String()
						if sofar != "" {
							searchterm.Reset()
							searchterm.WriteString(sofar[:len(sofar)-1])
						}
					}
					text = MakeTextImage(searchterm.String(), theme.Face, theme.Text)
					env.Draw() <- redraw(r, over, pressed, text, avatar, icon, info, cursor, phtext, isOpen, showCursor)
				}

			case win.MoMove:
				over = e.Point.In(r)
				env.Draw() <- redraw(r, over, pressed, text, avatar, icon, info, cursor, phtext, isOpen, showCursor)

			case win.MoDown:
				newPressed := e.Point.In(r)
				if newPressed != pressed {
					pressed = newPressed
					env.Draw() <- redraw(r, over, pressed, text, avatar, icon, info, cursor, phtext, isOpen, showCursor)
				}

			case win.MoUp:
				if pressed {
					if e.Point.In(r) {
						st := searchterm.String()
						if isOpen {
							search(st)
							if st != "" {
								sentSearch = st
								go getAvatar(sentSearch)
							}
							searchterm.Reset()
							exitCursor <- struct{}{}
						} else {
							go animateCursor(exitCursor)
						}
						phtext = MakeTextImage(sentSearch, theme.Face, theme.Text)
						text = MakeTextImage("", theme.Face, theme.Text)
						isOpen = !isOpen
					}
					pressed = false
					env.Draw() <- redraw(r, over, pressed, text, avatar, icon, info, cursor, phtext, isOpen, showCursor)
				}

			case win.KbType:
				if isOpen && isAlphanumeric(e.Rune) {
					searchterm.WriteRune(e.Rune)
				}
				text = MakeTextImage(searchterm.String(), theme.Face, theme.Text)
				env.Draw() <- redraw(r, over, pressed, text, avatar, icon, info, cursor, phtext, isOpen, showCursor)
			}

		}
	}
}

const alphanumeric = "qwertyuiopasdfghjklzxcvbnm1234567890-_"

func isAlphanumeric(key rune) bool {
	return strings.ContainsRune(alphanumeric, key)
}
