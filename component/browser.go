package component

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"time"

	"github.com/juanefec/gui"
	"github.com/juanefec/gui/win"
	"github.com/juanefec/scplayer/sc"
	. "github.com/juanefec/scplayer/util"
	"golang.org/x/image/math/fixed"
)

type pnext struct {
	sc.Song
	i int
}

func Browser(env gui.Env, theme *Theme, action <-chan string, song2player chan<- sc.Song, move <-chan int, pausebtn chan<- bool, reloadUser <-chan string, newInfo chan<- string, gotop <-chan struct{}, gotosong <-chan struct{}, listenBrowserSlider <-chan int, newBrowserSlider chan<- int, updateBrowserSlider chan<- int, playingPos chan<- int) {
	username := "kr3a71ve"

	const inset = 4
	metrics := theme.Face.Metrics()
	lineHeight := (metrics.Height + 2*fixed.I(inset)).Ceil()

	createImage := func(images []image.Image) *image.RGBA {
		var width int
		for _, img := range images {
			if img.Bounds().Dx() > width {
				width = img.Bounds().Inset(-inset).Dx()
			}
		}

		height := lineHeight * len(images)

		namesImage := image.NewRGBA(image.Rect(0, 0, width+2*inset, height+2*inset))
		for i := range images {
			r := image.Rect(
				0, lineHeight*i,
				width, lineHeight*(i+1),
			)
			DrawLeftCentered(namesImage, r.Inset(inset), images[i], draw.Over)
		}

		return namesImage
	}

	reload := func(songs []sc.Song) *image.RGBA {
		var images []image.Image

		for _, song := range songs {
			images = append(images, MakeTextImage(fmt.Sprintf("[%v]    %v - %v", song.Username, song.Artist, song.Title), theme.Face, theme.Text))
		}
		return createImage(images)
	}

	reloadPlaynext := func(pnexts []pnext) *image.RGBA {
		var images []image.Image

		for _, ns := range pnexts {
			if ns.OriginalID != 0 {
				images = append(images, MakeTextImage(fmt.Sprintf("[%v]    %v - %v", ns.Username, ns.Artist, ns.Title), theme.Face, theme.Text))
			}
		}
		return createImage(images)
	}

	redraw := func(r image.Rectangle, showing string, selected int, position image.Point, lineHeight int, namesImage image.Image, playnextsView []pnext) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, &image.Uniform{theme.Background}, image.Point{}, draw.Src)
			draw.Draw(drw, r, namesImage, position, draw.Over)
			if selected >= 0 {
				highlightR := image.Rect(
					namesImage.Bounds().Min.X,
					namesImage.Bounds().Min.Y+lineHeight*selected,
					r.Max.X,
					namesImage.Bounds().Min.Y+lineHeight*(selected+1),
				)
				highlightR = highlightR.Sub(position).Add(r.Min)
				draw.DrawMask(
					drw, highlightR.Intersect(r),
					&image.Uniform{theme.Highlight}, image.Point{},
					&image.Uniform{color.Alpha{64}}, image.Point{},
					draw.Over,
				)
			}

			if len(playnextsView) > 0 {
				maskOn := func(on int) {
					highlightR := image.Rect(
						namesImage.Bounds().Min.X,
						namesImage.Bounds().Min.Y+lineHeight*on,
						r.Max.X,
						namesImage.Bounds().Min.Y+lineHeight*(on+1),
					)
					highlightR = highlightR.Sub(position).Add(r.Min)
					draw.DrawMask(
						drw, highlightR.Intersect(r),
						&image.Uniform{theme.NextHighlight}, image.Point{},
						&image.Uniform{color.Alpha{64}}, image.Point{},
						draw.Over,
					)
				}
				if showing != "playlist" {
					for _, pnx := range playnextsView {
						maskOn(pnx.i)
					}
				} else {
					for ind, _ := range playnextsView {
						maskOn(ind)
					}
				}
			}
			return r
		}
	}

	songs := []sc.Song{
		{Artist: "          type ENTER and open the searchbar             < - - -", Title: "- >    quit the searchbar with ESCAPE       "},
		{Artist: "search for a username (eg. soundcloud.com/<usename>)  < - - - -", Title: "- - > or if its empty just press ENTER      "},
		{Artist: "     type ENTER and fetch all the user likes.       < - - - - -", Title: "- - - >                                     "},
	}

	tracks := []sc.Song{}

	likes := []sc.Song{}

	var (
		err          error
		r            image.Rectangle
		selected     = -1
		selectedOGID = -1

		likesImage, tracksImage, playlistImage *image.RGBA
		namesImage                             = image.NewRGBA(r)

		showing  string = "tracks" // "likes" || "tracks" || "playlist"
		position        = map[string]image.Point{
			"tracks":   image.Point{},
			"likes":    image.Point{},
			"playlist": image.Point{},
		}

		playnexts            []pnext
		playnextsViewMatches []pnext
		mouseRbPressed       bool
	)

	tracksImage = reload(tracks)
	likesImage = reload(likes)
	playlistImage = reloadPlaynext(playnextsViewMatches)

	refresh := func(fetch bool) {
		playnextsViewMatches = []pnext{}
		if fetch {
			namesImage = image.NewRGBA(r)
			playlistImage = reloadPlaynext(playnextsViewMatches)
			env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)
			newInfo <- "loading..."
			tracks, err = sc.GetTracks(username)
			if err != nil {
				tracks = []sc.Song{
					{Artist: "- - - - - - -", Title: fmt.Sprintf("- - - - >  %v                 ", err.Error())},
				}
			}
			likes, err = sc.GetLikes(username)
			if err != nil {
				likes = []sc.Song{
					{Artist: "- - - - - - -", Title: fmt.Sprintf("- - - - >  %v                 ", err.Error())},
				}
			}
		}
		if showing == "tracks" {
			tracksImage = reload(tracks)

			//playnexts = []pnext{}
			selected = -1
			for i, s := range tracks {
				if s.OriginalID == selectedOGID {
					selected = i
				}
				for _, pnx := range playnexts {
					if pnx.OriginalID == s.OriginalID {
						pnx.i = i
						playnextsViewMatches = append(playnextsViewMatches, pnx)
					}
				}
			}
			songs = tracks
			namesImage = tracksImage
			newInfo <- fmt.Sprintf("   %d tracks.", len(tracks))
		} else if showing == "playlist" {
			selected = -1
			playnextsViewMatches = playnexts
			playlistImage = reloadPlaynext(playnextsViewMatches)
			songs = []sc.Song{}
			for _, pnx := range playnextsViewMatches {
				songs = append(songs, pnx.Song)
			}

			namesImage = playlistImage
			newInfo <- fmt.Sprintf("   %d next", len(playnexts))
		} else {
			likesImage = reload(likes)

			selected = -1
			for i, s := range likes {
				if s.OriginalID == selectedOGID {
					selected = i
				}
				for _, pnx := range playnexts {
					if pnx.OriginalID == s.OriginalID {
						pnx.i = i
						playnextsViewMatches = append(playnextsViewMatches, pnx)
					}
				}
			}
			songs = likes
			namesImage = likesImage
			newInfo <- fmt.Sprintf("   %d likes.", len(likes))
		}
		newBrowserSlider <- namesImage.Rect.Dy()
		updateBrowserSlider <- position[showing].Y
		playingPos <- lineHeight * selected
		env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)

	}

	scrollView := func(e win.MoScroll) {
		newP := position[showing].Sub(e.Point.Mul(62))
		if newP.X > namesImage.Bounds().Max.X-r.Dx() {
			newP.X = namesImage.Bounds().Max.X - r.Dx()
		}
		if newP.Y > namesImage.Bounds().Max.Y-r.Dy() {
			newP.Y = namesImage.Bounds().Max.Y - r.Dy()
		}
		if newP.X < namesImage.Bounds().Min.X {
			newP.X = namesImage.Bounds().Min.X
		}
		if newP.Y < namesImage.Bounds().Min.Y {
			newP.Y = namesImage.Bounds().Min.Y
		}
		if newP != position[showing] {
			position[showing] = newP
			env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)
			updateBrowserSlider <- position[showing].Y
		}
	}

	clickEvent := func(e win.MoUp, showing string) {
		if !e.Point.In(r) {
			return
		}

		if showing != "playlist" {
			click := e.Point.Sub(r.Min).Add(position[showing])
			i := click.Y / lineHeight
			if i < 0 || i >= len(songs) {
				return
			}
			switch e.Button {
			case win.ButtonLeft:
				if selected != i {
					selected = i
					if songs[selected].OriginalID == 0 {
						return
					}

					selectedOGID = songs[selected].OriginalID
					song2player <- songs[selected]
					pausebtn <- true
					env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)
					playingPos <- lineHeight * selected
				}
			case win.ButtonRight:
				if songs[i].OriginalID == 0 {
					return
				}

				delete := -1
				for ipn := 0; ipn < len(playnexts); ipn++ {
					if songs[i].OriginalID == playnexts[ipn].OriginalID {
						delete = ipn
					}
				}

				if delete == -1 {
					pnx := pnext{songs[i], i}

					playnexts = append(playnexts, pnx)
					playnextsViewMatches = append(playnextsViewMatches, pnx)

				} else {
					playnexts = append(playnexts[:delete], playnexts[delete+1:]...)
					for ipn := 0; ipn < len(playnextsViewMatches); ipn++ {
						if songs[i].OriginalID == playnextsViewMatches[ipn].OriginalID {
							playnextsViewMatches = append(playnextsViewMatches[:ipn], playnextsViewMatches[ipn+1:]...)
						}
					}
				}

				env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)

			}
		} else {
			click := e.Point.Sub(r.Min).Add(position[showing])
			i := click.Y / lineHeight
			if i < 0 || i >= len(playnexts) {
				return
			}
			switch e.Button {
			case win.ButtonLeft:
				pnx := playnexts[i]
				selectedOGID = pnx.OriginalID
				song2player <- pnx.Song
				pausebtn <- true
				env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)
				//playingPos <- lineHeight * selected
			case win.ButtonRight:
				if playnexts[i].OriginalID == 0 {
					return
				}

				playnexts = append(playnexts[:i], playnexts[i+1:]...)
				refresh(false)
				env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)
			}
		}
	}

	shuffle := func() {
		if showing != "playlist" {
			rand.Seed(time.Now().Unix())
			rand.Shuffle(len(songs), func(i, j int) { songs[i], songs[j] = songs[j], songs[i] })
			namesImage = reload(songs)

			position[showing] = image.Point{}
			selected = -1
			for i, s := range songs {
				if s.OriginalID == selectedOGID {
					selected = i
				}
				for pi, pnx := range playnexts {
					if s.OriginalID == pnx.OriginalID {
						playnexts[pi].i = i
					}
				}
			}
			yPos := selected * lineHeight
			if selected != -1 {
				position[showing] = position[showing].Add(image.Point{0, yPos})
			}
			playnextsViewMatches = playnexts
			playlistImage = reloadPlaynext(playnextsViewMatches)
			env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)
			updateBrowserSlider <- position[showing].Y
			playingPos <- lineHeight * selected
		}
	}

	moveSong := func(m int) {
		if m > 0 && len(playnexts) > 0 {
			s := playnexts[0]
			if s.i >= 0 && s.i < len(songs) {
				for i, sng := range songs {
					if s.OriginalID == sng.OriginalID {
						selected = i

					}
				}

			}
			selectedOGID = s.OriginalID
			playnexts = playnexts[1:]
			selectedOGID = s.OriginalID
			refresh(false)
			song2player <- s.Song
			pausebtn <- true
			playingPos <- lineHeight * selected
			env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)
			return
		}
		if selected >= -1 && selected < len(songs)+1 && len(songs) > 0 {
			if selected == 0 && m == -1 {
				return
			}

			selected = selected + m

			if selected < 0 || selected > len(songs)-1 {
				selected = -1
				return
			}
			if selected < 0 || songs[selected].OriginalID == 0 {
				return
			}
			selectedOGID = songs[selected].OriginalID
			song2player <- songs[selected]
			pausebtn <- true
			playingPos <- lineHeight * selected
			env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)
		}
	}

	sliderAction := func(y int) {
		position[showing] = image.Point{0, y}
		env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)
	}

	go2top := func() {
		position[showing] = image.Point{}
		updateBrowserSlider <- position[showing].Y
		env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)
	}

	go2currentsong := func() {
		yPos := selected * lineHeight
		if selected != -1 {
			position[showing] = image.Point{0, yPos}
		}
		env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)
		updateBrowserSlider <- position[showing].Y
	}

	showTracks := func() {
		showing = "tracks"
		refresh(false)
	}

	showLikes := func() {
		showing = "likes"
		refresh(false)
	}

	showPlaylist := func() {
		showing = "playlist"
		refresh(false)
	}

	loadNew := func(newuser string) {
		if newuser != "" {
			username = newuser
			refresh(true)
		}
	}

	mrbSelectPlaynext := func(e win.MoMove, showing string) {
		if mouseRbPressed && showing != "playlist" {
			if !e.Point.In(r) {
				return
			}
			click := e.Point.Sub(r.Min).Add(position[showing])
			i := click.Y / lineHeight
			if i < 0 || i >= len(songs) {
				return
			}
			if songs[i].OriginalID == 0 {
				return
			}

			if len(playnexts) > 0 {
				last := playnexts[len(playnexts)-1].i
				if i == last {
					return
				}
			}
			pnx := pnext{songs[i], i}

			playnexts = append(playnexts, pnx)
			playnextsViewMatches = append(playnextsViewMatches, pnx)

			env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)
		}
	}

	for {
		select {
		case y := <-listenBrowserSlider:
			sliderAction(y)
		case <-gotop:
			go2top()
		case <-gotosong:
			go2currentsong()
		case newuser := <-reloadUser:
			loadNew(newuser)
		case action := <-action:
			switch action {
			case "refresh":
				refresh(true)
			case "shuffle":
				shuffle()
			case "tracks":
				showTracks()
			case "likes":
				showLikes()
			case "playlist":
				showPlaylist()
			}
		case m := <-move:
			moveSong(m)
		case e, ok := <-env.Events():
			if !ok {
				close(env.Draw())
				return
			}
			switch e := e.(type) {
			case gui.Resize:
				r = e.Rectangle
				env.Draw() <- redraw(r, showing, selected, position[showing], lineHeight, namesImage, playnextsViewMatches)
			case win.MoUp:
				clickEvent(e, showing)
				if e.Button == win.ButtonRight {
					mouseRbPressed = false
				}
			case win.MoDown:
				if e.Button == win.ButtonRight {
					mouseRbPressed = true
				}
			case win.MoMove:
				mrbSelectPlaynext(e, showing)
			case win.MoScroll:
				scrollView(e)
			}
		}
	}
}
