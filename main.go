package main

import (
	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"
	"github.com/juanefec/scplayer/component"
	"github.com/juanefec/scplayer/sc"
	. "github.com/juanefec/scplayer/util"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/gofont/gobold"
)

func run() {

	face, err := TTFToFace(gobold.TTF, 18)
	if err != nil {
		panic(err)
	}

	theme := &Theme{
		Face: face,

		Title:      colornames.Steelblue,
		Background: colornames.Darkseagreen, //colornames.Azure,
		Empty:      colornames.Dodgerblue,   //colornames.Seagreen,
		Text:       colornames.Black,
		Highlight:  colornames.Blueviolet,
		ButtonUp:   colornames.Lightgrey,
		ButtonDown: colornames.Grey,
		ButtonOver: colornames.Darkgoldenrod,
	}

	w, err := win.New(win.Title("scplayer"), win.Size(1000, 600), win.Resizable())
	if err != nil {
		panic(err)
	}

	mux, env := gui.NewMux(w)

	reloadUser := make(chan string)

	action := make(chan string)
	move := make(chan int)

	song := make(chan sc.Song)
	pausebtn := make(chan bool)
	pausebtnstatus := make(chan bool)
	updateTitle := make(chan string)

	go component.Button(EvenVerticalMaxMinY(EvenHorizontal(mux.MakeEnv(), 0, 1, 10), 0, 1, 16, 0, 40), theme, "refresh", func() {
		action <- "refresh"
	})

	go component.Button(EvenVerticalMaxMinY(EvenHorizontal(mux.MakeEnv(), 1, 2, 10), 0, 1, 16, 0, 40), theme, "back", func() {
		move <- -1
	})

	go component.PauseButton(EvenVerticalMaxMinY(EvenHorizontal(mux.MakeEnv(), 2, 3, 10), 0, 1, 16, 0, 40), theme, pausebtnstatus, func(playing bool) {
		pausebtn <- playing
	})

	go component.Button(EvenVerticalMaxMinY(EvenHorizontal(mux.MakeEnv(), 3, 4, 10), 0, 1, 16, 0, 40), theme, "forward", func() {
		move <- 1
	})

	go component.Button(EvenVerticalMaxMinY(EvenHorizontal(mux.MakeEnv(), 4, 5, 10), 0, 1, 16, 0, 40), theme, "shuffle", func() {
		action <- "shuffle"
	})

	go component.Title(EvenVerticalMaxMinY(EvenHorizontal(mux.MakeEnv(), 5, 10, 10), 0, 1, 16, 0, 40), theme, updateTitle)

	go component.Player(EvenVerticalMaxMinY(EvenHorizontal(mux.MakeEnv(), 0, 1, 1), 1, 2, 16, 40, 80), theme, song, pausebtn, move, updateTitle)

	go component.Browser(EvenVerticalMaxMinY(EvenHorizontal(mux.MakeEnv(), 0, 1, 1), 2, 16, 16, 80, 1080), theme, action, song, move, pausebtnstatus, reloadUser)

	go component.Searchbar(EvenVertical(EvenHorizontal(mux.MakeEnv(), 1, 2, 3), 7, 8, 20), theme, func(searchterm string) {
		reloadUser <- searchterm
	})

	for e := range env.Events() {
		switch e.(type) {
		case win.WiClose:
			close(env.Draw())
		}
	}
}

func main() {
	mainthread.Run(run)
}
