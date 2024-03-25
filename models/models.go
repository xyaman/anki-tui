package models

import (
	"path/filepath"
	"strings"
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
	CardID int      `json:"cardId"`
	Fields Fields   `json:"fields"`
	Tags   []string `json:"tags"`

	// custom fields
	sentenceValue string
	morphsValue   string
	audioValue    string
	imageValue    string
}

// Fields represents the main fields for a Anki Note
type Fields map[string]interface{}

func (n *Note) GetFieldsValues(sentence, morphs, audio, image string) {
	sentenceFieldsName := strings.Split(sentence, ",")
	for _, fieldName := range sentenceFieldsName {
		if sentenceField, ok := n.Fields[fieldName]; ok {
			n.sentenceValue = sentenceField.(map[string]interface{})["value"].(string)
			break
		}
	}

	morphsFieldsName := strings.Split(morphs, ",")
	for _, fieldName := range morphsFieldsName {
		if morphsField, ok := n.Fields[fieldName]; ok {
			n.morphsValue = morphsField.(map[string]interface{})["value"].(string)
			break
		}
	}

	audioFieldsName := strings.Split(audio, ",")
	for _, fieldName := range audioFieldsName {
		if audioField, ok := n.Fields[fieldName]; ok {
			n.audioValue = audioField.(map[string]interface{})["value"].(string)
			break
		}
	}

	imageFieldsName := strings.Split(image, ",")
	for _, fieldName := range imageFieldsName {
		if imageField, ok := n.Fields[fieldName]; ok {
			n.imageValue = imageField.(map[string]interface{})["value"].(string)
			break
		}
	}
}

func (n *Note) GetSentence() string {
	return n.sentenceValue
}

func (n *Note) GetMorphs() string {
	return n.morphsValue
}

func (n *Note) GetAudioValue() string {
	return n.audioValue
}

func (n *Note) GetAudioPath(mediaCollection string) string {
	audioValue := n.GetAudioValue()
	if audioValue == "" {
		return ""
	}
	return filepath.Join(mediaCollection, audioValue[7:len(audioValue)-1])
}

func (n *Note) GetImageValue() string {
	return n.imageValue
}

func (n *Note) GetImagePath(mediaCollection string) string {
	imageValue := n.GetImageValue()
	if imageValue == "" {
		return ""
	}
	return filepath.Join(mediaCollection, imageValue[10:len(imageValue)-2])
}
