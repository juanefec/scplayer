package component

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/juanefec/gui"
	"github.com/juanefec/gui/win"
	"github.com/juanefec/scplayer/sc"
	. "github.com/juanefec/scplayer/util"
)

func SelectDynamic(mux *gui.Mux, minI, maxI, n, minY, maxY, startx int, theme *Theme, listeningTime <-chan string, action func(sc.User), option <-chan sc.User, swapOption <-chan sc.User) {

	var (
		click     = make(chan sc.User)
		newXStart = make(chan int)
	)

	width := 26

	optionMap := map[string]opt{}

	i := startx
	ni := 0

	addOption := func(op sc.User) {
		optionMap[op.Username] = opt{ni, make(chan sc.User), make(chan struct{}), make(chan [2]int)}
		go SelectButtonDynamic(
			FixedFromTopLeft(mux.MakeEnv(), i, i+width, minY, maxY),
			theme, op, optionMap[op.Username],
			func(opName sc.User) {
				click <- opName
			})
		i += width
		newXStart <- i
		ni++
	}

	go Infobar(EvenVerticalMinMaxY(mux.MakeEnv(), minI, maxI, n, minY, maxY), theme, listeningTime, newXStart)
	newXStart <- i
	lastClick := ""
	for {
		select {
		case nop := <-option:
			if nop.Err != nil {
				continue
			}
			if _, ok := optionMap[nop.Username]; !ok {
				addOption(nop)
			}
		case nop := <-swapOption:
			if nop.Err != nil {
				optionMap[lastClick].exit <- struct{}{}
				i -= width
				newXStart <- i
				sn := optionMap[lastClick].n
				for _, v := range optionMap {
					if v.n > sn {
						v.n--
						v.newX <- [2]int{width, width}
					}
				}
				delete(optionMap, lastClick)
				continue
			}
			optionMap[lastClick].in <- nop
			optionMap[nop.Username] = optionMap[lastClick]
			delete(optionMap, lastClick)
		case ns := <-click:
			lastClick = ns.Username
			action(ns)
		}
	}
}

type opt struct {
	n    int
	in   chan sc.User
	exit chan struct{}
	newX chan [2]int
}

func SelectButtonDynamic(env gui.Env, theme *Theme, user sc.User, com opt, action func(sc.User)) {
	redraw := func(r image.Rectangle, icon image.Image, over, pressed bool) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			var clr color.Color
			if pressed {
				clr = theme.TextBoxUp
			} else if over {
				clr = theme.TextBoxOver
			} else {
				clr = theme.Title
			}
			draw.Draw(drw, r, &image.Uniform{clr}, image.Point{}, draw.Src)
			DrawCentered(drw, r, icon, draw.Over)
			return r
		}
	}

	var (
		r         image.Rectangle
		over      bool
		pressed   bool
		avatarImg image.Image = user.AvSmol
	)

	for {
		select {
		case <-com.exit:
			fmt.Println("exiting ", user.Username)
			close(env.Draw())
			return
		case nu := <-com.in:
			user = nu
			avatarImg = user.AvSmol
			env.Draw() <- redraw(r, avatarImg, over, pressed)
		case nu := <-com.newX:
			r.Min.X, r.Max.X = r.Min.X-nu[0], r.Max.X-nu[1]
			env.Draw() <- redraw(r, avatarImg, over, pressed)
		case e := <-env.Events():
			switch e := e.(type) {
			case win.MoMove:
				over = e.Point.In(r)
				env.Draw() <- redraw(r, avatarImg, over, pressed)
			case gui.Resize:
				r = e.Rectangle
				env.Draw() <- redraw(r, avatarImg, over, pressed)

			case win.MoDown:
				newPressed := e.Point.In(r)
				if newPressed != pressed {
					pressed = newPressed
					env.Draw() <- redraw(r, avatarImg, over, pressed)
				}

			case win.MoUp:
				if pressed {
					if e.Point.In(r) {
						action(user)
					}
					pressed = false
					env.Draw() <- redraw(r, avatarImg, over, pressed)
				}
			}
		}
	}
}
