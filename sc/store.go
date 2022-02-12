package sc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
)

const (
	MUSIC_DIR  = "songs"
	AUDIO_FILE = "audio.mp3"
	COVER_FILE = "cover.jpeg"
	DESC_FILE  = "desc.json"
)

var (
	PWD = ""
)

func init() {
	err := error(nil)
	PWD, err = os.Getwd()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func LocalCheck(originalid int) (*bytes.Buffer, image.Image, bool) {
	id := fmt.Sprintf("%v/%v/%v", PWD, MUSIC_DIR, originalid)
	audioPath := fmt.Sprintf("%v/%v", id, AUDIO_FILE)
	coverPath := fmt.Sprintf("%v/%v", id, COVER_FILE)
	var (
		err         error
		audioBytes  []byte
		coverBytes  []byte
		audioBuffer = &bytes.Buffer{}
		cover       image.Image
	)
	if audioBytes, err = ioutil.ReadFile(audioPath); err != nil {
		return nil, nil, false
	}

	if coverBytes, err = ioutil.ReadFile(coverPath); err == nil {
		cover, err = jpeg.Decode(bytes.NewReader(coverBytes))
		if err != nil {
			fmt.Println(err)

		}
	}

	_, err = audioBuffer.Write(audioBytes)
	if err != nil {
		return nil, nil, false
	}

	return audioBuffer, cover, true
}

func LocalSave(song SongData, audio []byte, cover image.Image) error {
	id := fmt.Sprintf("%v/%v/%v", PWD, MUSIC_DIR, song.OriginalID)
	audioPath := fmt.Sprintf("%v/%v", id, AUDIO_FILE)
	coverPath := fmt.Sprintf("%v/%v", id, COVER_FILE)
	descPath := fmt.Sprintf("%v/%v", id, DESC_FILE)
	fmt.Println(audioPath)

	err := os.MkdirAll(id, 0777)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	err = os.WriteFile(audioPath, audio, 0777)
	if err != nil {
		return err
	}

	if cover != nil {
		coverFile, err := os.Create(coverPath)
		if err != nil {
			return err
		}
		err = jpeg.Encode(coverFile, cover, nil)
		if err != nil {
			return err
		}
		coverFile.Close()
	}

	rawdata, err := json.Marshal(song)
	if err != nil {
		return err
	}

	err = os.WriteFile(descPath, rawdata, 0777)

	if err != nil {
		return err
	}

	return nil
}

func LoadLocals() []Song {
	localsongs := fmt.Sprintf("%v/%v", PWD, MUSIC_DIR)

	songs := []Song{}

	localSongs, err := os.Open(localsongs)
	if err != nil {
		return songs
	}

	sids, err := localSongs.Readdirnames(0)
	if err != nil {
		return songs
	}

	for _, id := range sids {
		songPath := fmt.Sprintf("%v/%v", localsongs, id)
		descPath := fmt.Sprintf("%v/%v", songPath, DESC_FILE)

		descBytes, err := ioutil.ReadFile(descPath)
		if err != nil {
			continue
		}

		sd := &SongData{}
		err = json.Unmarshal(descBytes, sd)
		if err != nil {
			continue
		}
		songs = append(songs, sd.AsSong())
	}
	return songs
}
