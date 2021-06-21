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

		Title:        colornames.Steelblue,
		Background:   colornames.Darkseagreen, //colornames.Azure,
		Empty:        colornames.Dodgerblue,   //colornames.Seagreen,
		Text:         colornames.Black,
		Highlight:    colornames.Blueviolet,
		ButtonUp:     colornames.Lightgrey,
		ButtonDown:   colornames.Grey,
		ButtonOver:   colornames.Darkgoldenrod,
		VolumeBg:     color.RGBA{0x50, 0x50, 0x50, 0xff},
		VolumeBgOver: color.RGBA{0x60, 0x60, 0x60, 0xff},
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
	newInfo := make(chan string)

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

	go component.VolumeSlider(EvenVerticalMinMaxY(EvenHorizontalMinMaxX(mux.MakeEnv(), 0, 1, 14, 0, 60), 1, 2, 16, 40, 80), theme, trigUpdateVolume, func(v float64) {
		updateVolume <- v
	})

	go component.Player(EvenVerticalMinMaxY(EvenHorizontalMinMaxX(mux.MakeEnv(), 1, 14, 14, 60, 1920), 1, 2, 16, 40, 80), theme, song, pausebtn, move, updateTitle, updateVolume, trigUpdateVolume)

	go component.Infobar(EvenVerticalMinMaxY(EvenHorizontal(mux.MakeEnv(), 0, 1, 1), 2, 3, 16, 80, 100), theme, newInfo, func(searchterm string) {
		reloadUser <- searchterm
	})

	go component.Browser(EvenVerticalMinMaxY(EvenHorizontal(mux.MakeEnv(), 0, 1, 1), 3, 16, 16, 100, 1080), theme, action, song, move, pausebtnstatus, reloadUser, newInfo)

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
