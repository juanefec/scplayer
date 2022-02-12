package component

import (
	"fmt"
	"image"
	"image/draw"
	"time"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	"github.com/juanefec/scplayer/sc"

	. "github.com/juanefec/scplayer/util"
)

func Player(env gui.Env, theme *Theme, newsong <-chan sc.Song, pausebtn <-chan bool, next chan<- int, updateTitle chan<- NewTitle, updateVol <-chan float64, listeningTime chan<- string) {

	redraw := func(r image.Rectangle, rail image.Image, progress image.Image, imgProgress image.Image, imgProgressTop image.Image) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, &image.Uniform{theme.Empty}, image.Point{}, draw.Src)
			DrawLeftCentered(drw, r.Inset(4), imgProgress, draw.Over)
			DrawRightCentered(drw, r.Inset(4), imgProgressTop, draw.Over)

			if rail != nil {
				DrawCentered(drw, r.Inset(4), rail, draw.Over)
			}

			if progress != nil {
				DrawCentered(drw, r.Inset(4), progress, draw.Over)
			}

			return r
		}
	}

	var (
		r               image.Rectangle
		rRail           image.Rectangle
		imgProgress     image.Image
		imgProgressTop  image.Image
		progress        image.Image
		rail            image.Image
		exitSongStarter = make(chan struct{})
		done            = make(chan struct{})
		doneDownload    = make(chan int)
		doneTimer       = make(chan struct{})
		playing         bool

		pvol float64

		playingTime time.Duration

		song *sc.Song = &sc.Song{}
	)

	loadSong := func(nsong sc.Song) {
		song = &nsong
		updateTitle <- NewTitle{Title: "loading..."}
		go func(song *sc.Song) {
			err := song.Download(doneDownload)
			if err != nil {
				fmt.Println(err)
			}
		}(song)

		sc.ClearSpeaker()
	}

	songEnd := func() {
		playing = false
		doneTimer <- struct{}{}
		next <- 1
	}

	pauseEvent := func(bplaying bool) {
		if bplaying {
			playing = false
			song.Pause()
		} else {
			err := song.Resume()
			if err != nil {
				next <- 1
			} else {
				playing = true
			}
		}
	}

	songTimer := func() {
		for {
			select {
			case <-doneTimer:
				return

			case <-time.After(time.Millisecond * 100):

				imgProgress = MakeTextImage(song.Progress(), theme.Face, theme.Text)
				imgProgressTop = MakeTextImage(song.Duration(), theme.Face, theme.Text)
				rRail = r
				rRail.Min.X, rRail.Max.X = r.Min.X+34, r.Max.X-50
				if ra, pr, ok := MakeRailAndProgressImage(rRail, song.DurationMs(), song.ProgressMs(), theme.Rail); ok {
					rail, progress = ra, pr
				}

				env.Draw() <- redraw(r, rail, progress, imgProgress, imgProgressTop)
			}
		}
	}

	songStarter := func(exit chan struct{}) {
		for {
			select {
			case <-exit:
				doneTimer <- struct{}{}
				return
			case <-time.After(time.Second):
				if playing {
					playingTime += time.Second
					listeningTime <- sc.DurationToStr(playingTime)
				}
			case dd := <-doneDownload:
				if dd == song.OriginalID {
					if playing {
						doneTimer <- struct{}{}
					}

					err := song.Play(pvol, done)
					if err != nil {
						fmt.Println(err)
						continue
					}

					playing = true
					title := fmt.Sprintf("%v - %v", song.Artist, song.Title)

					updateTitle <- NewTitle{Title: title, Cover: song.Cover}
					go songTimer()
				}
			}
		}
	}

	eventLoop := func() {
		for e := range env.Events() {
			switch e := e.(type) {
			case gui.Resize:
				r = e.Rectangle
				env.Draw() <- redraw(r, rail, progress, imgProgress, imgProgressTop)
			case win.MoDown:
				if song.IsDownloaded() && e.Button == win.ButtonLeft {
					if e.Point.In(rRail) {
						rawPos := e.Point.X - rRail.Min.X
						pos := Map(rawPos, 0, rRail.Dx(), 0, 100)
						song.SetProgress(pos)
					}
				}
			}
		}
		exitSongStarter <- struct{}{}
		close(env.Draw())
	}

	go eventLoop()
	go songStarter(exitSongStarter)

	for {
		select {

		case <-done:
			songEnd()
		case bplaying := <-pausebtn:
			pauseEvent(bplaying)
		case nsong := <-newsong:
			loadSong(nsong)
		case v := <-updateVol:
			pvol = song.SetVolume(v)

		}
	}
}

type NewTitle struct {
	Title string
	Cover image.Image
}

func Title(env gui.Env, theme *Theme, newTitle <-chan NewTitle) {
	redraw := func(r image.Rectangle, imgTitle, imgCover image.Image) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, &image.Uniform{theme.Title}, image.Point{}, draw.Src)
			DrawCentered(drw, r, imgTitle, draw.Over)
			DrawLeftCentered(drw, r, imgCover, draw.Over)
			return r
		}
	}

	var (
		r        image.Rectangle
		imgTitle image.Image
		imgCover image.Image
	)

	for {
		select {

		case t := <-newTitle:
			imgTitle = MakeTextScaledImage(t.Title, theme.Face, theme.Text, 1.5)
			imgCover = t.Cover
			env.Draw() <- redraw(r, imgTitle, imgCover)
		case e, ok := <-env.Events():
			if !ok {
				close(env.Draw())
				return
			}
			if resize, ok := e.(gui.Resize); ok {
				r = resize.Rectangle
				env.Draw() <- redraw(r, imgTitle, imgCover)
			}
		}
	}
}
