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

func Browser(env gui.Env, theme *Theme, action <-chan string, song2player chan<- sc.Song, move <-chan int, pausebtn chan<- bool, reloadUser <-chan string, newInfo chan<- string, gotop <-chan struct{}, gotosong <-chan struct{}, listenBrowserSlider <-chan int, newBrowserSlider chan<- int, updateBrowserSlider chan<- int, playingPos chan<- int) {
	username := "kr3a71ve"

	reload := func(songs []sc.Song) ([]sc.Song, int, *image.RGBA) {

		var images []image.Image
		//fmt.Printf("Listed %v songs.\n", len(songs))
		for _, song := range songs {
			images = append(images, MakeTextImage(fmt.Sprintf("%v - %v", song.Artist, song.Title), theme.Face, theme.Text))
		}

		const inset = 4

		var width int
		for _, img := range images {
			if img.Bounds().Dx() > width {
				width = img.Bounds().Inset(-inset).Dx()
			}
		}

		metrics := theme.Face.Metrics()
		lineHeight := (metrics.Height + 2*fixed.I(inset)).Ceil()
		height := lineHeight * len(songs)

		namesImage := image.NewRGBA(image.Rect(0, 0, width+2*inset, height+2*inset))
		for i := range images {
			r := image.Rect(
				0, lineHeight*i,
				width, lineHeight*(i+1),
			)
			DrawLeftCentered(namesImage, r.Inset(inset), images[i], draw.Over)
		}

		return songs, lineHeight, namesImage
	}

	type pnext struct {
		i   int
		ogi int
	}

	reloadPlaynext := func(songs []sc.Song, pnexts []pnext) *image.RGBA {
		var images []image.Image
		for _, s := range songs {
			for _, ns := range pnexts {
				if s.OriginalID == ns.ogi {
					images = append(images, MakeTextImage(fmt.Sprintf("%v - %v", s.Artist, s.Title), theme.Face, theme.Text))
				}
			}
		}
		const inset = 4

		var width int
		for _, img := range images {
			if img.Bounds().Dx() > width {
				width = img.Bounds().Inset(-inset).Dx()
			}
		}

		metrics := theme.Face.Metrics()
		lineHeight := (metrics.Height + 2*fixed.I(inset)).Ceil()
		height := lineHeight * len(pnexts)

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

	redraw := func(r image.Rectangle, selected int, position, positionPlaynext image.Point, lineHeight int, namesImage, playnextNamesImage image.Image, playnexts []pnext, showPlaynext bool) func(draw.Image) image.Rectangle {
		if showPlaynext {
			return func(drw draw.Image) image.Rectangle {
				draw.Draw(drw, r, &image.Uniform{theme.Background}, image.Point{}, draw.Src)
				draw.Draw(drw, r, playnextNamesImage, positionPlaynext, draw.Over)

				if len(playnexts) > 0 {
					for i, _ := range playnexts {
						highlightR := image.Rect(
							playnextNamesImage.Bounds().Min.X,
							playnextNamesImage.Bounds().Min.Y+lineHeight*i,
							r.Max.X,
							playnextNamesImage.Bounds().Min.Y+lineHeight*(i+1),
						)
						highlightR = highlightR.Sub(positionPlaynext).Add(r.Min)
						draw.DrawMask(
							drw, highlightR.Intersect(r),
							&image.Uniform{theme.NextHighlight}, image.Point{},
							&image.Uniform{color.Alpha{64}}, image.Point{},
							draw.Over,
						)
					}
				}
				return r
			}
		}
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
			if len(playnexts) > 0 {
				for _, pnx := range playnexts {
					highlightR := image.Rect(
						namesImage.Bounds().Min.X,
						namesImage.Bounds().Min.Y+lineHeight*pnx.i,
						r.Max.X,
						namesImage.Bounds().Min.Y+lineHeight*(pnx.i+1),
					)
					highlightR = highlightR.Sub(position).Add(r.Min)
					draw.DrawMask(
						drw, highlightR.Intersect(r),
						&image.Uniform{theme.NextHighlight}, image.Point{},
						&image.Uniform{color.Alpha{64}}, image.Point{},
						draw.Over,
					)
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

	var (
		err              error
		r                image.Rectangle
		position         = image.Point{}
		positionPlaynext = image.Point{}
		selected         = -1
		selectedOGID     = -1

		userResource string // "likes" || "tracks"

		playnexts    []pnext
		showPlaynext bool
	)

	songs, lineHeight, namesImage := reload(songs)
	playnextsImage := reloadPlaynext(songs, playnexts)

	refresh := func() {

		loadInfo := ""
		if userResource == "tracks" {
			songs, err = sc.GetTracks(username)
			if err != nil {
				songs = []sc.Song{
					{Artist: "- - - - - - -", Title: fmt.Sprintf("- - - - >  %v                 ", err.Error())},
				}
			}
			loadInfo = fmt.Sprintf("%v   %d tracks.", username, len(songs))
		} else {
			songs, err = sc.GetLikes(username)
			if err != nil {
				songs = []sc.Song{
					{Artist: "- - - - - - -", Title: fmt.Sprintf("- - - - >  %v                 ", err.Error())},
				}
			}
			loadInfo = fmt.Sprintf("%v   %d likes.", username, len(songs))
		}

		songs, lineHeight, namesImage = reload(songs)
		position = image.Point{}
		playnexts = []pnext{}
		selected = -1
		for i, s := range songs {
			if s.OriginalID == selectedOGID {
				selected = i
			}
		}
		newInfo <- loadInfo
		newBrowserSlider <- namesImage.Rect.Dy()
		env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)
		updateBrowserSlider <- position.Y
		playingPos <- lineHeight * selected
	}

	switchView := func() {
		if !showPlaynext {
			playnextsImage = reloadPlaynext(songs, playnexts)
			newBrowserSlider <- playnextsImage.Rect.Dy()
			updateBrowserSlider <- positionPlaynext.Y
		} else {
			songs, lineHeight, namesImage = reload(songs)
			newBrowserSlider <- namesImage.Rect.Dy()
			updateBrowserSlider <- position.Y
		}
		showPlaynext = !showPlaynext

		env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)
	}

	scrollView := func(e win.MoScroll) {
		if !showPlaynext {

			newP := position.Sub(e.Point.Mul(62))
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
			if newP != position {
				position = newP
				env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)
				updateBrowserSlider <- position.Y
			}
		} else {
			newP := positionPlaynext.Sub(e.Point.Mul(62))
			if newP.X > playnextsImage.Bounds().Max.X-r.Dx() {
				newP.X = playnextsImage.Bounds().Max.X - r.Dx()
			}
			if newP.Y > playnextsImage.Bounds().Max.Y-r.Dy() {
				newP.Y = playnextsImage.Bounds().Max.Y - r.Dy()
			}
			if newP.X < playnextsImage.Bounds().Min.X {
				newP.X = playnextsImage.Bounds().Min.X
			}
			if newP.Y < playnextsImage.Bounds().Min.Y {
				newP.Y = playnextsImage.Bounds().Min.Y
			}
			if newP != positionPlaynext {
				positionPlaynext = newP
				env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)
				updateBrowserSlider <- positionPlaynext.Y
			}
		}
	}

	clickEvent := func(e win.MoDown) {
		if !e.Point.In(r) {
			return
		}

		if !showPlaynext {
			click := e.Point.Sub(r.Min).Add(position)
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
					env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)
					playingPos <- lineHeight * selected
				}
			case win.ButtonRight:
				if songs[i].OriginalID == 0 {
					return
				}

				delete := -1
				for ipn := 0; ipn < len(playnexts); ipn++ {
					if songs[i].OriginalID == playnexts[ipn].ogi {
						delete = ipn
					}
				}

				if delete == -1 {
					playnexts = append(playnexts, pnext{i: i, ogi: songs[i].OriginalID})
				} else {
					playnexts = append(playnexts[:delete], playnexts[delete+1:]...)
				}

				env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)

			}
		} else {
			click := e.Point.Sub(r.Min).Add(positionPlaynext)
			i := click.Y / lineHeight
			if i < 0 || i >= len(playnexts) {
				return
			}
			switch e.Button {
			case win.ButtonLeft:
				pnx := playnexts[i]
				if pnx.i < 0 || pnx.i >= len(songs) || songs[pnx.i].OriginalID == 0 {
					return
				}

				selectedOGID = songs[pnx.i].OriginalID
				song2player <- songs[pnx.i]
				pausebtn <- true
				env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)
				//playingPos <- lineHeight * selected
			}
		}
	}

	shuffle := func() {
		rand.Seed(time.Now().Unix())
		rand.Shuffle(len(songs), func(i, j int) { songs[i], songs[j] = songs[j], songs[i] })
		songs, lineHeight, namesImage = reload(songs)

		position = image.Point{}
		selected = -1
		for i, s := range songs {
			if s.OriginalID == selectedOGID {
				selected = i
			}
			for pi, pnx := range playnexts {
				if s.OriginalID == pnx.ogi {
					playnexts[pi].i = i
				}
			}
		}
		yPos := selected * lineHeight
		if selected != -1 {
			position = position.Add(image.Point{0, yPos})
		}

		env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)
		updateBrowserSlider <- position.Y
		playingPos <- lineHeight * selected
	}

	moveSong := func(m int) {
		if selected >= -1 && selected < len(songs)+1 && len(songs) > 0 {
			if selected == 0 && m == -1 {
				return
			}
			if m > 0 {
				if len(playnexts) > 0 {
					selected = playnexts[0].i
					selectedOGID = playnexts[0].ogi
					playnexts = playnexts[1:]
					playnextsImage = reloadPlaynext(songs, playnexts)

				} else {
					selected = selected + m
				}
			} else {
				selected = selected + m
			}
			if selected < 0 && selected > len(songs)-1 {
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
			env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)
		}
	}

	sliderAction := func(y int) {
		if !showPlaynext {
			position = image.Point{0, y}
		} else {
			positionPlaynext = image.Point{0, y}
		}
		env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)
	}

	go2top := func() {
		if !showPlaynext {
			position = image.Point{}
			updateBrowserSlider <- position.Y
		} else {
			positionPlaynext = image.Point{}
			updateBrowserSlider <- positionPlaynext.Y
		}
		env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)
	}

	go2currentsong := func() {
		yPos := selected * lineHeight
		if selected != -1 {
			position = image.Point{0, yPos}
		}
		env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)
		updateBrowserSlider <- position.Y

	}

	loadNew := func(newuser string) {
		if newuser != "" {
			username = newuser
			refresh()
		}
		env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)
	}

	tracks := func() {
		userResource = "tracks"
		refresh()
	}

	likes := func() {
		userResource = "likes"
		refresh()
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
				refresh()
			case "shuffle":
				shuffle()
			case "tracks":
				tracks()
			case "likes":
				likes()
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
				env.Draw() <- redraw(r, selected, position, positionPlaynext, lineHeight, namesImage, playnextsImage, playnexts, showPlaynext)
			case win.MoDown:
				clickEvent(e)
			case win.MoScroll:
				scrollView(e)
			case win.KbDown:
				switch e.Key {
				case win.KeyTab:
					switchView()
				}
				// case win.KbUp:
				// 	switch e.Key {
				// 	case win.KeySpace:
				// 	}
			}
		}
	}
}
