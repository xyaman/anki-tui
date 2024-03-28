package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/xyaman/anki-tui/models"
)

type ExternalSource interface {
	FetchNotesFromQuery(query string, start, end int) ([]models.Note, error)
}

type BrigadaSource struct {
	http   http.Client
	apiKey string
}

type BrigadaSOSResponse struct {
	Sentences []Sentence `json:"sentences"`
}

type Sentence struct {
	BasicInfo   BasicInfo   `json:"basic_info"`
	SegmentInfo SegmentInfo `json:"segment_info"`
	MediaInfo   MediaInfo   `json:"media_info"`
}

type BasicInfo struct {
	NameAnimeJp string `json:"name_anime_jp"`
	NameAnimeEn string `json:"name_anime_en"`
}

type SegmentInfo struct {
	ContentJp string `json:"content_jp"`
	IsNsfw    bool   `json:"is_nsfw"`
	ActorJa   string `json:"actor_ja"`
	ActorEn   string `json:"actor_en"`
	ActorEs   string `json:"actor_es"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type MediaInfo struct {
	PathImage string `json:"path_image"`
	PathAudio string `json:"path_audio"`
	PathVideo string `json:"path_video"`
}

func NewBrigadaSource(apiKey string) *BrigadaSource {
	return &BrigadaSource{
		http:   http.Client{},
		apiKey: apiKey,
	}
}

func (b *BrigadaSource) FetchNotesFromQuery(query string, start, end int) ([]models.Note, error) {

	jsonBody := []byte(fmt.Sprintf(`{"query":"%s","exact_match":0,"limit":%d,"content_sort":null,"random_seed":null,"season":null,"episode":null}`, query, end-start+1))

	req, err := http.NewRequest("POST", "https://api.brigadasos.xyz/api/v1/api/search/anime/sentence", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", b.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := b.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	var parsedResponse BrigadaSOSResponse
	err = json.NewDecoder(res.Body).Decode(&parsedResponse)
	if err != nil {
		return nil, err
	}

	notes := make([]models.Note, len(parsedResponse.Sentences))
	for i, sentence := range parsedResponse.Sentences {

		starttime := strings.ReplaceAll(sentence.SegmentInfo.StartTime, ":", "_")
		starttime = strings.ReplaceAll(starttime, ".", "_")
		endtime := strings.ReplaceAll(sentence.SegmentInfo.EndTime, ":", "_")
		endtime = strings.ReplaceAll(endtime, ".", "_")

		name := strings.ReplaceAll(sentence.BasicInfo.NameAnimeEn, " ", "_")
		name = strings.ReplaceAll(name, "/", "")
		name = strings.ReplaceAll(name, ":", "_")
		name = strings.ReplaceAll(name, "'", "")
		name = strings.ReplaceAll(name, "!", "")
		name = strings.ReplaceAll(name, "?", "")
		name = strings.ReplaceAll(name, ".", "")
		name = strings.ReplaceAll(name, ",", "")
		name = strings.ReplaceAll(name, "(", "")
		name = strings.ReplaceAll(name, ")", "")
		name = strings.ReplaceAll(name, "\\", "")

		notes[i] = models.Note{
			NoteID:        i,
			SentenceValue: sentence.SegmentInfo.ContentJp,
			AudioValue:    sentence.MediaInfo.PathAudio,
			ImageValue:    sentence.MediaInfo.PathImage,
			Tags:          []string{sentence.BasicInfo.NameAnimeJp},
			Source:        "BrigadaSOS",
			Filename:      fmt.Sprintf("%s_%s_%s", name, starttime, endtime),
		}
	}

	// Fetch images concurrently and add them to the notes
	// wait until is complete to return the function
	wg := sync.WaitGroup{}
	for i := range notes {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			notes[i].GetImage("")
		}(i)
	}
	wg.Wait()

	return notes, nil
}
