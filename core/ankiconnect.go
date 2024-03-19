package core

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

  "github.com/xyaman/anki-tui/models"
)

// AnkiConnect is the Client for the AnkiConnect API
type AnkiConnect struct {
	Url        string
	Version    int
	httpClient *http.Client
}

func NewAnkiConnect(url string, version int) *AnkiConnect {
	return &AnkiConnect{
		Url:        url,
		Version:    version,
		httpClient: &http.Client{},
	}
}

// request is a helper function to make a request to the AnkiConnect API, the response needs to be unmarshalled by the caller
func (c *AnkiConnect) request(action string, params interface{}) ([]byte, error) {

	requestBody, err := json.Marshal(map[string]interface{}{
		"action":  action,
		"params":  params,
		"version": 6,
	})

	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.Url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *AnkiConnect) FindNotes(query string) (*models.FindNotesResult, error) {
	result, err := c.request("findNotes", map[string]interface{}{
		"query": query,
	})
	if err != nil {
		return nil, err
	}

	var notes *models.FindNotesResult
	err = json.Unmarshal(result, &notes)
	if err != nil {
		return nil, err
	}

	return notes, nil
}

func (c *AnkiConnect) NotesInfo(notes []int) (*models.NotesInfoResult, error) {
	result, err := c.request("notesInfo", map[string]interface{}{
		"notes": notes,
	})
	if err != nil {
		return nil, err
	}

	var notesInfo *models.NotesInfoResult
	err = json.Unmarshal(result, &notesInfo)
	if err != nil {
		return nil, err
	}

	return notesInfo, nil
}

func (c *AnkiConnect) DeleteNotes(cards []int) error {
	_, err := c.request("deleteNotes", map[string]interface{}{
		"notes": cards,
	})
	return err
}

func (c *AnkiConnect) UpdateNoteFields(noteID int, fields models.Fields) error {
	_, err := c.request("updateNoteFields", map[string]interface{}{
		"note": map[string]interface{}{
			"id":     noteID,
			"fields": fields,
		},
	})
	return err
}

func (c *AnkiConnect) AddTags(noteID int, tag string) error {
	_, err := c.request("addTags", map[string]interface{}{
		"notes": []int{noteID},
		"tags":  tag,
	})

	return err
}

func (c *AnkiConnect) GetLastAddedCard() (*models.Note, error) {
	lastNotes, err := c.FindNotes("added:2")
	if err != nil {
		return nil, err
	}

	lastNoteID := 0
	for _, currentNoteId := range lastNotes.Result {
		if currentNoteId > lastNoteID {
			lastNoteID = currentNoteId
		}
	}

	// Get Note Info
	lastNote, err := c.NotesInfo([]int{lastNoteID})
	if err != nil {
		return nil, err
	}

	return &lastNote.Result[0], nil
}


func (c *AnkiConnect) GuiBrowse(query string) error {
	_, err := c.request("guiBrowse", map[string]interface{}{
		"query": query,
	})

	return err
}

func (c *AnkiConnect) GetMediaDirPath() (string, error) {
	result, err := c.request("getMediaDirPath", map[string]interface{}{})
	if err != nil {
		return "", err
	}

	var mediaDirPathResponse map[string]interface{}
	err = json.Unmarshal(result, &mediaDirPathResponse)
	if err != nil {
		return "", err
	}

	mediaDirPath, ok := mediaDirPathResponse["result"].(string)
	if !ok {
		return "", err
	}
	return mediaDirPath, nil
}
