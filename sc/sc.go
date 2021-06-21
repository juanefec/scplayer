package sc

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/juanefec/scplayer/util"
	scp "github.com/zackradisic/soundcloud-api"
)

func GetLikes(username string) ([]Song, error) {
	if username == "" {
		return nil, fmt.Errorf("empty username")
	}

	sc, err := scp.New(scp.APIOptions{})

	if err != nil {
		log.Fatal(err.Error())
	}

	user, err := sc.GetUser(scp.GetUserOptions{
		ProfileURL: "https://soundcloud.com/" + username,
	})

	if err != nil {
		return nil, fmt.Errorf("user [%v] not found", username)
		//log.Fatal(err.Error())
	}

	return getAllLikes(sc, user.ID, 0), nil

}

func getAllLikes(sc *scp.API, user int64, offset int) []Song {
	ls, err := sc.GetLikes(scp.GetLikesOptions{
		ID:     user,
		Type:   "track",
		Limit:  200,
		Offset: offset,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	l, err := ls.GetLikes()
	if err != nil {
		log.Fatal(err.Error())
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
	// 		log.Fatal(err.Error())
	// 	}

	// 	off, err := strconv.Atoi(url.Query()["offset"][0])
	// 	if err != nil {
	// 		log.Fatal(err.Error())
	// 	}

	// 	songs = append(songs, getAllLikes(sc, user, off)...)
	// }

	return songs
}

type Song struct {
	Title      string
	Artist     string
	OriginalID int
	duration   time.Duration
	data       scp.Transcoding
	volume     *effects.Volume
	controller *beep.Ctrl
	streamer   beep.StreamSeekCloser
	format     beep.Format
}

func (song *Song) Play(done chan<- struct{}) error {
	sc, err := scp.New(scp.APIOptions{})

	if err != nil {
		log.Fatal(err.Error())
	}

	buffer := &bytes.Buffer{}

	// its a prop song to write on the browser probably.
	if song.OriginalID == 0 {
		done <- struct{}{}
		return nil
	}

	err = sc.DownloadTrack(song.data, buffer)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}

	streamer, format, err := mp3.Decode(ioutil.NopCloser(buffer))
	if err != nil {
		log.Fatal(err)
		return err
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	ctrl := &beep.Ctrl{
		Streamer: beep.Seq(streamer, beep.Callback(func() {
			buffer.Reset()
			done <- struct{}{}
		})),
		Paused: false,
	}

	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   -4,
		Silent:   false,
	}

	song.format = format
	song.streamer = streamer
	song.controller = ctrl
	song.volume = volume
	speaker.Play(song.volume)

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

func (song Song) Resume() error {
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
