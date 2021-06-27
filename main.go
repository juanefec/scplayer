package main

import (
	"image"
	"image/color"

	"github.com/faiface/mainthread"
	"github.com/juanefec/gui"
	"github.com/juanefec/gui/win"
	"github.com/juanefec/scplayer/component"
	"github.com/juanefec/scplayer/icons"
	"github.com/juanefec/scplayer/sc"
	. "github.com/juanefec/scplayer/util"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

func run() {

	// ttffile, err := os.ReadFile("pxfont.TTF")
	// if err != nil {
	// 	panic(err)
	// }
	// face, err := TTFToFace(ttffile, 16)
	// if err != nil {
	// 	panic(err)
	// }

	theme := &Theme{
		Face: basicfont.Face7x13,

		Title:            color.RGBA{0x4B, 0x10, 0x4B, 0xff}, //color.RGBA{0x1C, 0x5C, 0x2C, 0xff}, //colornames.Steelblue,
		Background:       color.RGBA{0x17, 0x0F, 0x11, 0xff}, //colornames.Azure,
		Empty:            color.RGBA{0x17, 0x0F, 0x11, 0xff}, //colornames.Dodgerblue,              //colornames.Seagreen,
		Text:             colornames.Whitesmoke,
		Rail:             colornames.Whitesmoke,
		NextHighlight:    color.RGBA{0x1C, 0x5C, 0x2C, 0xff},
		Highlight:        color.RGBA{0xc2, 0x11, 0xc5, 0xff},
		HighlightSlider:  color.RGBA{0xc2, 0x11, 0xc5, 0xff},
		BackgroundSlider: color.RGBA{0x82, 0x55, 0x84, 0xff},
		Infobar:          color.RGBA{0x4B, 0x10, 0x4B, 0xff}, // //color.RGBA{0x55, 0x10, 0x56, 0xff},
		ButtonUp:         color.RGBA{0x82, 0x55, 0x84, 0xff},
		ButtonDown:       color.RGBA{0x89, 0x70, 0x8f, 0xff},
		ButtonOver:       color.RGBA{0xAA, 0x98, 0xAE, 0xff},
		TextBoxUp:        color.RGBA{0x3D, 0x10, 0x3B, 0xff},
		TextBoxDown:      color.RGBA{0x89, 0x70, 0x8f, 0xff},
		TextBoxOver:      color.RGBA{0x5B, 0x20, 0x5B, 0xff},
		VolumeBg:         color.RGBA{0x17, 0x0F, 0x11, 0xff},
		VolumeBgOver:     color.RGBA{0x17, 0x0F, 0x11, 0xff},
	}

	appIcon := icons.GetIcon("app-icon")

	w, err := win.New(win.Title("scplayer"), win.Size(1000, 600), win.Resizable(), win.Icon([]image.Image{appIcon}))
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
	gotop := make(chan struct{})
	gotosong := make(chan struct{})
	newInfo := make(chan string)
	listenBrowser := make(chan int)
	updateBrowser := make(chan int)
	listenNewBrowser := make(chan int)
	playingPos := make(chan int)

	listeningTime := make(chan string)

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

	go component.Title(EvenVerticalMinMaxY(EvenHorizontalMinX(mux.MakeEnv(), 5, 10, 10, 200), 0, 1, 16, 0, 40), theme, updateTitle)

	go component.VolumeSlider(EvenVerticalMinMaxY(EvenHorizontalMinMaxX(mux.MakeEnv(), 0, 1, 14, 0, 52), 1, 2, 16, 40, 90), theme, func(v float64) {
		updateVolume <- v
	})

	go component.Player(EvenVerticalMinMaxY(EvenHorizontalMinX(mux.MakeEnv(), 1, 14, 14, 52), 1, 2, 16, 40, 90), theme, song, pausebtn, move, updateTitle, updateVolume, listeningTime)

	go component.Searchbar(EvenVerticalMinMaxY(FixedFromLeft(mux.MakeEnv(), 0, 200), 2, 3, 16, 90, 114), theme, func(searchterm string) {
		reloadUser <- searchterm
	})

	go component.Infobar(EvenVerticalMinMaxY(EvenHorizontalRightMinMaxX(mux.MakeEnv(), 0, 1, 1, 200, 52), 2, 3, 16, 90, 114), theme, newInfo, listeningTime, func(searchterm string) {
		reloadUser <- searchterm
	})

	go component.Button(EvenVerticalMinMaxY(FixedFromRight(mux.MakeEnv(), 0, 26), 2, 3, 16, 90, 114), theme, "gotop", func() {
		gotop <- struct{}{}
	})

	go component.Button(EvenVerticalMinMaxY(FixedFromRight(mux.MakeEnv(), 26, 52), 2, 3, 16, 90, 114), theme, "gotosong", func() {
		gotosong <- struct{}{}
	})

	go component.SelectGeneric(mux, 2, 3, 16, 114, 138, theme, func(tab string) {
		action <- tab
	}, "tracks", "tracks", "likes", "playlist")

	go component.Browser(EvenVerticalMinMaxY(EvenHorizontalRightMinX(mux.MakeEnv(), 0, 1, 1, 52), 3, 16, 16, 138, 1080), theme, action, song, move, pausebtnstatus, reloadUser, newInfo, gotop, gotosong, updateBrowser, listenNewBrowser, listenBrowser, playingPos)

	go component.BrowserSlider(EvenVerticalMinMaxY(FixedFromRight(mux.MakeEnv(), 0, 52), 3, 16, 16, 138, 1080), theme, listenBrowser, updateBrowser, listenNewBrowser, playingPos)

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
