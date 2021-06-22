package main

import (
	"image/color"

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

		Title:           color.RGBA{0x4B, 0x10, 0x4B, 0xff}, //color.RGBA{0x1C, 0x5C, 0x2C, 0xff}, //colornames.Steelblue,
		Background:      color.RGBA{0x17, 0x0F, 0x11, 0xff}, //colornames.Azure,
		Empty:           color.RGBA{0x17, 0x0F, 0x11, 0xff}, //colornames.Dodgerblue,              //colornames.Seagreen,
		Text:            colornames.Whitesmoke,
		Rail:            colornames.Whitesmoke,
		NextHighlight:   color.RGBA{0x1C, 0x5C, 0x2C, 0xff},
		Highlight:       color.RGBA{0xc2, 0x11, 0xc5, 0xff},
		HighlightSlider: color.RGBA{0x4E, 0x67, 0x66, 0xff},
		Infobar:         color.RGBA{0x4E, 0x67, 0x66, 0xff}, // //color.RGBA{0x55, 0x10, 0x56, 0xff},
		ButtonUp:        color.RGBA{0x82, 0x55, 0x84, 0xff},
		ButtonDown:      color.RGBA{0x89, 0x70, 0x8f, 0xff},
		ButtonOver:      color.RGBA{0xAA, 0x98, 0xAE, 0xff},
		VolumeBg:        color.RGBA{0x17, 0x0F, 0x11, 0xff},
		VolumeBgOver:    color.RGBA{0x17, 0x0F, 0x11, 0xff},
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
	updateVolume := make(chan float64)
	trigUpdateVolume := make(chan struct{})
	gotop := make(chan struct{})
	gotosong := make(chan struct{})
	newInfo := make(chan string)
	listenBrowser := make(chan int)
	updateBrowser := make(chan int)
	listenNewBrowser := make(chan int)
	playingPos := make(chan int)

	go component.Button(EvenVerticalMinMaxY(EvenHorizontalMinMaxX(mux.MakeEnv(), 0, 1, 10, 0, 40), 0, 1, 16, 0, 40), theme, "refresh", func() {
		action <- "refresh"
	})

	go component.Button(EvenVerticalMinMaxY(EvenHorizontalMinMaxX(mux.MakeEnv(), 1, 2, 10, 40, 80), 0, 1, 16, 0, 40), theme, "back", func() {
		move <- -1
	})

	go component.PauseButton(EvenVerticalMinMaxY(EvenHorizontalMinMaxX(mux.MakeEnv(), 2, 3, 10, 80, 120), 0, 1, 16, 0, 40), theme, pausebtnstatus, func(playing bool) {
		pausebtn <- playing
	})

	go component.Button(EvenVerticalMinMaxY(EvenHorizontalMinMaxX(mux.MakeEnv(), 3, 4, 10, 120, 160), 0, 1, 16, 0, 40), theme, "forward", func() {
		move <- 1
	})

	go component.Button(EvenVerticalMinMaxY(EvenHorizontalMinMaxX(mux.MakeEnv(), 4, 5, 10, 160, 200), 0, 1, 16, 0, 40), theme, "shuffle", func() {
		action <- "shuffle"
	})

	go component.Title(EvenVerticalMinMaxY(EvenHorizontalMinMaxX(mux.MakeEnv(), 5, 10, 10, 200, 1920), 0, 1, 16, 0, 40), theme, updateTitle)

	go component.VolumeSlider(EvenVerticalMinMaxY(EvenHorizontalMinMaxX(mux.MakeEnv(), 0, 1, 14, 0, 60), 1, 2, 16, 40, 90), theme, trigUpdateVolume, func(v float64) {
		updateVolume <- v
	})

	go component.Player(EvenVerticalMinMaxY(EvenHorizontalMinX(mux.MakeEnv(), 1, 14, 14, 60), 1, 2, 16, 40, 90), theme, song, pausebtn, move, updateTitle, updateVolume, trigUpdateVolume)

	go component.Button(EvenVerticalMinMaxY(FixedFromRight(mux.MakeEnv(), 0, 26), 2, 3, 16, 90, 114), theme, "gotop", func() {
		gotop <- struct{}{}
	})

	go component.Button(EvenVerticalMinMaxY(FixedFromRight(mux.MakeEnv(), 26, 52), 2, 3, 16, 90, 114), theme, "gotosong", func() {
		gotosong <- struct{}{}
	})

	go component.Infobar(EvenVerticalMinMaxY(EvenHorizontalRightMinX(mux.MakeEnv(), 0, 1, 1, 52), 2, 3, 16, 90, 114), theme, newInfo, func(searchterm string) {
		reloadUser <- searchterm
	})

	go component.Browser(EvenVerticalMinMaxY(EvenHorizontalRightMinX(mux.MakeEnv(), 0, 1, 1, 52), 3, 16, 16, 114, 1080), theme, action, song, move, pausebtnstatus, reloadUser, newInfo, gotop, gotosong, updateBrowser, listenNewBrowser, listenBrowser, playingPos)

	go component.BrowserSlider(EvenVerticalMinMaxY(FixedFromRight(mux.MakeEnv(), 0, 52), 3, 16, 16, 114, 1080), theme, listenBrowser, updateBrowser, listenNewBrowser, playingPos)

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
