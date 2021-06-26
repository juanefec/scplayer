package sc

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/juanefec/scplayer/util"
	scp "github.com/juanefec/soundcloud-api"
)

var url = "https://api-v2.soundcloud.com/stream/users/188669904?offset=2021-03-31T06%3A19%3A56.000Z%2Ctracks%2C01019580628&limit=10&client_id=ahAJuiWvqPHUWMtUhizqN5QaITxmOwTN&app_version=1624267273&app_locale=es"

var url2 = "https://api-v2.soundcloud.com/users/188669904/tracks?representation=&client_id=ahAJuiWvqPHUWMtUhizqN5QaITxmOwTN&limit=20&offset=0&linked_partitioning=1&app_version=1624267273&app_locale=es"

func GetTracks(username string) ([]Song, error) {
	if username == "" {
		return nil, fmt.Errorf("empty username")
	}

	sc, err := scp.New(scp.APIOptions{})

	if err != nil {
		// log.Fatal(err.Error())
		return nil, err
	}

	user, err := sc.GetUser(scp.GetUserOptions{
		ProfileURL: "https://soundcloud.com/" + username,
	})

	if err != nil {
		return nil, fmt.Errorf("user [%v] not found", username)
		//// log.Fatal(err.Error())
		//return err
	}

	return getAllTracks(sc, user.ID, 0)

}

func getAllTracks(sc *scp.API, user int64, offset int) ([]Song, error) {
	ls, err := sc.GetTracklist(scp.GetTracklistOptions{
		ID:     user,
		Type:   "tracklist",
		Limit:  200,
		Offset: offset,
	})
	if err != nil {
		// log.Fatal(err.Error())
		return nil, err
	}

	tks, err := ls.GetTracks()
	if err != nil {
		// log.Fatal(err.Error())
		return nil, err
	}

	songs := []Song{}
	for _, trk := range tks {
		if len(trk.Media.Transcodings) > 0 {
			s := Song{
				Title:      trk.Title,
				Artist:     trk.User.Username,
				OriginalID: int(trk.ID),
				duration:   time.Duration(trk.FullDurationMS * 1000000),
				data:       trk.Media.Transcodings[0],
			}

			songs = append(songs, s)
		}
	}

	// Recursion disabled for developing, it takes some time to bring all tracks sometimes.
	//
	// After mapping SC Tracks to Songs we look for the sc_response.NextHref, this prop
	// contains the url that follows your search, this way you will be able to retrive all
	// the list.
	// Here I just take the offset value form the url.Query() and pass it recusively

	// if ls.NextHref != "" {
	// 	url, err := url.Parse(ls.NextHref)
	// 	if err != nil {
	// 		// log.Fatal(err.Error())
	//		return nil, err
	// 	}

	// 	off, err := strconv.Atoi(url.Query()["offset"][0])
	// 	if err != nil {
	// 		// log.Fatal(err.Error())
	//		return nil, err
	// 	}

	// 	songs = append(songs, getAllLikes(sc, user, off)...)
	// }

	return songs, nil
}

func GetLikes(username string) ([]Song, error) {
	if username == "" {
		return nil, fmt.Errorf("empty username")
	}

	sc, err := scp.New(scp.APIOptions{})

	if err != nil {
		// log.Fatal(err.Error())
		return nil, err
	}

	user, err := sc.GetUser(scp.GetUserOptions{
		ProfileURL: "https://soundcloud.com/" + username,
	})

	if err != nil {
		return nil, fmt.Errorf("user [%v] not found", username)
		//// log.Fatal(err.Error())
		//return err
	}

	return getAllLikes(sc, user.ID, 0)

}

func getAllLikes(sc *scp.API, user int64, offset int) ([]Song, error) {
	ls, err := sc.GetTracklist(scp.GetTracklistOptions{
		ID:     user,
		Type:   "track",
		Limit:  200,
		Offset: offset,
	})
	if err != nil {
		// log.Fatal(err.Error())
		return nil, err
	}

	l, err := ls.GetLikes()
	if err != nil {
		// log.Fatal(err.Error())
		return nil, err
	}

	songs := []Song{}
	for _, li := range l {
		if len(li.Track.Media.Transcodings) > 0 {
			s := Song{
				Title:      li.Track.Title,
				Artist:     li.Track.User.Username,
				OriginalID: int(li.Track.ID),
				duration:   time.Duration(li.Track.FullDurationMS * 1000000),
				data:       li.Track.Media.Transcodings[0],
			}

			songs = append(songs, s)
		}
	}

	// Recursion disabled for developing, it takes some time to bring all tracks sometimes.
	//
	// After mapping SC Tracks to Songs we look for the sc_response.NextHref, this prop
	// contains the url that follows your search, this way you will be able to retrive all
	// the list.
	// Here I just take the offset value form the url.Query() and pass it recusively

	// if ls.NextHref != "" {
	// 	url, err := url.Parse(ls.NextHref)
	// 	if err != nil {
	// 		// log.Fatal(err.Error())
	//		return nil, err
	// 	}

	// 	off, err := strconv.Atoi(url.Query()["offset"][0])
	// 	if err != nil {
	// 		// log.Fatal(err.Error())
	//		return nil, err
	// 	}

	// 	songs = append(songs, getAllLikes(sc, user, off)...)
	// }

	return songs, nil
}

type Song struct {
	Title        string
	Artist       string
	OriginalID   int
	duration     time.Duration
	data         scp.Transcoding
	volume       *effects.Volume
	controller   *beep.Ctrl
	streamer     beep.StreamSeekCloser
	format       beep.Format
	buffer       *bytes.Buffer
	isDownloaded bool
}

func (song *Song) Play(vol float64, done chan<- struct{}) error {

	speaker.Init(song.format.SampleRate, song.format.SampleRate.N(time.Second/10))

	ctrl := &beep.Ctrl{
		Streamer: beep.Seq(song.streamer, beep.Callback(func() {
			song.buffer.Reset()
			done <- struct{}{}
		})),
		Paused: false,
	}

	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   vol,
		Silent:   false,
	}

	song.controller = ctrl
	song.volume = volume
	speaker.Play(song.volume)

	return nil
}

func (song *Song) Download(done chan<- int) error {
	// its a prop song to write on the browser probably.
	if song.OriginalID == 0 {
		//done <- 0
		return fmt.Errorf("error")
	}

	sc, err := scp.New(scp.APIOptions{})

	if err != nil {
		// log.Fatal(err.Error())
		return err
	}

	buffer := &bytes.Buffer{}

	err = sc.DownloadTrack(song.data, buffer)
	if err != nil {
		// log.Fatal(err.Error())
		return err
		//return err
	}

	streamer, format, err := mp3.Decode(ioutil.NopCloser(buffer))
	if err != nil {
		//log.Fatal(err)
		return err
	}

	song.format = format
	song.streamer = streamer
	song.buffer = buffer
	song.isDownloaded = true
	done <- song.OriginalID
	return nil
}

func (song *Song) SetVolume(vol float64) {
	if song == nil || song.volume == nil {
		return
	}
	if vol < 0.3 {
		song.volume.Silent = true
		return
	}
	song.volume.Silent = false
	rs, re := -6.5, 1.8
	song.volume.Volume = util.MapFloat(float64(vol), 0, 9, rs, re)
}

func (song *Song) GetVolume() float64 {
	if song == nil || song.volume == nil {
		return -4.0
	}
	return song.volume.Volume
}

func (song Song) Progress() string {
	if song.streamer == nil {
		return ""
	}
	speaker.Lock()
	t := song.format.SampleRate.D(song.streamer.Position())
	speaker.Unlock()

	return durationToStr(t)
}

func (song Song) Duration() string {
	return durationToStr(song.duration)
}

func (song Song) ProgressMs() int {
	if song.streamer == nil {
		return 0
	}
	speaker.Lock()
	t := song.format.SampleRate.D(song.streamer.Position())
	speaker.Unlock()

	return int(t / time.Millisecond)
}

func (song Song) DurationMs() int {
	return int(song.duration / time.Millisecond)
}

func (song *Song) Resume() error {
	if song.controller == nil {
		return errors.New("not playing")
	}
	song.controller.Paused = false
	return nil
}

func (song Song) Pause() {
	if song.controller != nil {
		song.controller.Paused = true
	}
}

func (song Song) Stop() {
	speaker.Clear()
}

func ClearSpeaker() {
	speaker.Clear()
}

func durationToStr(d time.Duration) string {
	h, m, s := 0, 0, 0
	{
		d = d.Round(time.Second)
		hd := d / time.Hour
		d -= hd * time.Hour
		md := d / time.Minute
		d -= md * time.Minute
		sd := d / time.Second
		h, m, s = int(hd), int(md), int(sd)
	}
	if h >= 1 {
		return fmt.Sprintf("%v:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
