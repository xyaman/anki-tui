package models

import (
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"golang.org/x/image/webp"
)

type FindNotesResult struct {
	Result []int  `json:"result"`
	Error  string `json:"error"`
}

type NotesInfoResult struct {
	Result []Note `json:"result"`
	Error  string `json:"error"`
}

type Note struct {
	NoteID int      `json:"noteId"`
	Fields Fields   `json:"fields"`
	Tags   []string `json:"tags"`

	// custom fields
	Source        string
	SentenceValue string
	MorphsValue   string
	AudioValue    string
	ImageValue    string

	Image    image.Image
	Filename string
}

// Fields represents the main fields for a Anki Note
type Fields map[string]interface{}

func (n *Note) GetFieldsValues(sentence, morphs, audio, image string) {
	sentenceFieldsName := strings.Split(sentence, ",")
	for _, fieldName := range sentenceFieldsName {
		if sentenceField, ok := n.Fields[fieldName]; ok {
			n.SentenceValue = sentenceField.(map[string]interface{})["value"].(string)
			break
		}
	}

	morphsFieldsName := strings.Split(morphs, ",")
	for _, fieldName := range morphsFieldsName {
		if morphsField, ok := n.Fields[fieldName]; ok {
			n.MorphsValue = morphsField.(map[string]interface{})["value"].(string)
			break
		}
	}

	audioFieldsName := strings.Split(audio, ",")
	for _, fieldName := range audioFieldsName {
		if audioField, ok := n.Fields[fieldName]; ok {
			n.AudioValue = audioField.(map[string]interface{})["value"].(string)
			break
		}
	}

	imageFieldsName := strings.Split(image, ",")
	for _, fieldName := range imageFieldsName {
		if imageField, ok := n.Fields[fieldName]; ok {
			n.ImageValue = imageField.(map[string]interface{})["value"].(string)
			break
		}
	}
}

func (n *Note) GetSource() string {
	if n.Source == "" {
		return "Anki"
	}
	return n.Source
}

func (n *Note) GetSentence() string {
	return n.SentenceValue
}

func (n *Note) GetMorphs() string {
	return n.MorphsValue
}

func (n *Note) GetAudioValue() string {
	return n.AudioValue
}

func (n *Note) GetAudio(mediaCollection string) (io.ReadCloser, beep.StreamSeekCloser) {
	if n.GetSource() == "BrigadaSOS" {
		res, err := http.Get(n.GetAudioValue())
		if err != nil {
			panic(err)
		}

		// defer res.Body.Close()
		streamer, _, err := mp3.Decode(res.Body)
		if err != nil {
			panic(err)
		}

		return res.Body, streamer
	}

	if n.GetAudioValue() == "" {
		return nil, nil
	}

	audioPath := n.GetAudioValue()[7 : len(n.GetAudioValue())-1]
	reader, err := os.Open(filepath.Clean(filepath.Join(mediaCollection, audioPath)))
	if err != nil {
		panic(err)
	}

	// defer reader.Close()
	streamer, _, err := mp3.Decode(reader)
	if err != nil {
		panic(err)
	}
	return reader, streamer
}

func (n *Note) GetImageValue() string {
	return n.ImageValue
}

func (n *Note) GetImage(mediaCollection string) image.Image {
	if n.Image != nil {
		return n.Image
	}

	// TODO: Add this method to the interface
	if n.GetSource() == "BrigadaSOS" {
		res, err := http.Get(n.GetImageValue())
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
		img, err := webp.Decode(res.Body)
		if err != nil {
			panic(err)
		}

		n.Image = img
		return img
	}

	imageValue := n.GetImageValue()
	if imageValue == "" {
		return nil
	}

	imageContent, err := os.Open(filepath.Clean(filepath.Join(mediaCollection, imageValue[10:len(imageValue)-2])))
	if err != nil {
		panic(err)
	}

	img, _, err := image.Decode(imageContent)
	if err != nil {
		panic(err)
	}

	return img
}

func (n *Note) GetFilename() string {
	return n.Filename
}

// DownloadImage saves the image to the media collection and returns the path
func (n *Note) DownloadImage(collectionPath string) (string, error) {

	var imagePath string
	// TODO implement this method in the interface
	if n.GetSource() == "BrigadaSOS" {

		res, err := http.Get(n.GetImageValue())
		if err != nil {
			return "", err
		}

		defer res.Body.Close()
		imagePath = filepath.Join(collectionPath, n.GetFilename()+".webp")
		file, err := os.Create(imagePath)
		if err != nil {
			return "", err
		}

		_, err = io.Copy(file, res.Body)
		if err != nil {
			return "", err
		}
	}

	// Anki field format: <img src="image.jpg">
	imageFieldValue := fmt.Sprintf("<img src=\"%s\">", filepath.Base(imagePath))
	return imageFieldValue, nil
}

func (n *Note) DownloadAudio(collectionPath string) (string, error) {
	if n.GetAudioValue() == "" {
		return "", errors.New("Note audio is nil")
	}

	var audioPath string
	// TODO implement this method in the interface
	if n.GetSource() == "BrigadaSOS" {
		res, err := http.Get(n.GetAudioValue())
		if err != nil {
			return "", err
		}

		defer res.Body.Close()
		audioPath = filepath.Join(collectionPath, n.GetFilename()+".mp3")
		file, err := os.Create(audioPath)
		if err != nil {
			return "", err
		}

		_, err = io.Copy(file, res.Body)
		if err != nil {
			return "", err
		}
	}

	// Anki field format: [sound:audio.mp3]
	audioFieldValue := fmt.Sprintf("[sound:%s]", filepath.Base(audioPath))
	return audioFieldValue, nil
}
