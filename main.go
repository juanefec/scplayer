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

	var (
		pausebtnstatus    = make(chan bool)               // pauseButton 	 	  <--- browser
		updateCurrentUser = make(chan string)             // searchBar	     	  <--- userSelect
		startedLoading    = make(chan struct{})           // avatar				  <--- searchBar
		listeningTime     = make(chan string)             // userSelect(infoBar)  <--- player
		addUserHistory    = make(chan sc.User)            // userSelect			  <--- browser
		updateTitle       = make(chan component.NewTitle) // title	              <--- player
		reloadUserAvatar  = make(chan sc.User)            // avatar               <--- userSelect, browser
		updateVolume      = make(chan float64)            // player				  <--- volume
		song              = make(chan sc.Song)            // player				  <--- browser
		updateBrowser     = make(chan int)                // browser 		 	  <--- browserSlider
		listenNewBrowser  = make(chan int)                // browser 		 	  <--- browserSlider
		pausebtn          = make(chan bool)               // browser		 	  <--- pauseButton
		reloadUser        = make(chan string)             // browser	     	  <--- searchBar <-userSelect
		action            = make(chan string)             // browser	     	  <--- refreshButton <-shuffleButton <-userSelect
		clearPlaylist     = make(chan struct{})           // browser			  <--- buttonClearPlaylist xd
		swapOption        = make(chan sc.User)            // browser              <--- userSelect
		move              = make(chan int)                // browser 		 	  <--- buttonBack <-buttonForward <-player (when song ends)
		listenBrowser     = make(chan int)                // browserSlider   	  <--- browser
		playingPos        = make(chan int)                // browserSlider   	  <--- browser
	)

	go component.Button(FixedFromTopLeft(mux.MakeEnv(), 0, 40, 0, 40), theme, "refresh", func() {
		action <- "refresh"
	})

	go component.Button(FixedFromTopLeft(mux.MakeEnv(), 40, 80, 0, 40), theme, "back", func() {
		move <- -1
	})

	go component.PauseButton(FixedFromTopLeft(mux.MakeEnv(), 80, 120, 0, 40), theme, pausebtnstatus, func(playing bool) {
		pausebtn <- playing
	})

	go component.Button(FixedFromTopLeft(mux.MakeEnv(), 120, 160, 0, 40), theme, "forward", func() {
		move <- 1
	})

	go component.Button(FixedFromTopLeft(mux.MakeEnv(), 160, 200, 0, 40), theme, "shuffle", func() {
		action <- "shuffle"
	})

	go component.Title(FixedFromTopXBounds(mux.MakeEnv(), 200, 0, 0, 40), theme, updateTitle)

	go component.VolumeSlider(FixedFromTopLeft(mux.MakeEnv(), 0, 52, 40, 90), theme, func(v float64) {
		updateVolume <- v
	})

	go component.Player(FixedFromTopXBounds(mux.MakeEnv(), 52, 0, 40, 90), theme, song, pausebtn, move, updateTitle, updateVolume, listeningTime)

	go component.Avatar(FixedFromTopLeft(mux.MakeEnv(), 0, 52, 90, 142), theme, reloadUserAvatar, startedLoading)

	go component.Searchbar(FixedFromTopLeft(mux.MakeEnv(), 52, 200, 90, 116), theme, func(searchterm string) {
		reloadUser <- searchterm
		if searchterm != "" {
			startedLoading <- struct{}{}
		}
	}, updateCurrentUser)

	go component.SelectDynamic(mux, 2, 3, 16, 90, 116, 200, theme, listeningTime, func(user sc.User) {
		reloadUser <- user.Username
		reloadUserAvatar <- user
		updateCurrentUser <- user.Username
	}, addUserHistory, swapOption)

	// go component.Infobar(EvenVerticalMinMaxY(FixedFromRight(mux.MakeEnv(), 0, 124), 2, 3, 16, 90, 116), theme, newInfo, listeningTime, func(searchterm string) {
	// 	reloadUser <- searchterm
	// })

	go component.Button(EvenVerticalMinMaxY(FixedFromRight(mux.MakeEnv(), 0, 52), 2, 3, 16, 116, 142), theme, "clear-playlist", func() {
		clearPlaylist <- struct{}{}
	})

	go component.SelectGeneric(mux, 2, 3, 16, 116, 142, 52, 52, theme, func(tab string) {
		action <- tab
	}, "tracks", "tracks", "likes", "playlist")

	go component.Browser(EvenVerticalMinMaxY(FixedFromBounds(mux.MakeEnv(), 0, 52), 3, 16, 16, 142, 1080), theme,
		action,
		song,
		move,
		pausebtnstatus,
		reloadUser,
		updateBrowser,
		listenNewBrowser,
		listenBrowser,
		playingPos,
		clearPlaylist,
		addUserHistory,
		reloadUserAvatar,
		swapOption,
	)

	go component.BrowserSlider(EvenVerticalMinMaxY(FixedFromRight(mux.MakeEnv(), 0, 52), 3, 16, 16, 142, 1080), theme, listenBrowser, updateBrowser, listenNewBrowser, playingPos)

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
