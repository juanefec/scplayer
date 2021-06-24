package component

import (
	"fmt"
	"image"
	"image/draw"
	"time"

	"github.com/juanefec/gui"
	"github.com/juanefec/scplayer/sc"

	. "github.com/juanefec/scplayer/util"
)

func Player(env gui.Env, theme *Theme, newsong <-chan sc.Song, pausebtn <-chan bool, next chan<- int, updateTitle chan<- string, updateVol <-chan float64) {

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

		song *sc.Song = &sc.Song{}
	)

	loadSong := func(nsong sc.Song) {
		pvol = song.GetVolume()
		song = &nsong

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
			case _, ok := <-env.Events():
				if !ok {
					return
				}
			case <-doneTimer:
				return

			case <-time.After(time.Millisecond * 100):

				imgProgress = MakeTextImage(song.Progress(), theme.Face, theme.Text)
				imgProgressTop = MakeTextImage(song.Duration(), theme.Face, theme.Text)

				if ra, pr, ok := MakeRailAndProgressImage(r, song.DurationMs(), song.ProgressMs(), theme.Rail); ok {
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
				return
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

					updateTitle <- title
					go songTimer()
				}
			}
		}
	}

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
			song.SetVolume(v)
		case e, ok := <-env.Events():
			if !ok {
				close(env.Draw())
				exitSongStarter <- struct{}{}
				return
			}
			switch e := e.(type) {
			case gui.Resize:
				r = e.Rectangle
				env.Draw() <- redraw(r, rail, progress, imgProgress, imgProgressTop)
			}
		}
	}
}

func Title(env gui.Env, theme *Theme, newTitle <-chan string) {
	redraw := func(r image.Rectangle, imgTitle image.Image) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, &image.Uniform{theme.Title}, image.Point{}, draw.Src)
			DrawCentered(drw, r, imgTitle, draw.Over)
			return r
		}
	}

	var (
		r        image.Rectangle
		imgTitle image.Image
	)

	for {
		select {
		case t := <-newTitle:
			imgTitle = MakeTextScaledImage(t, theme.Face, theme.Text, 1.5)
			env.Draw() <- redraw(r, imgTitle)
		case e, ok := <-env.Events():
			if !ok {
				close(env.Draw())
				return
			}
			if resize, ok := e.(gui.Resize); ok {
				r = resize.Rectangle
				env.Draw() <- redraw(r, imgTitle)
			}
		}
	}
}
