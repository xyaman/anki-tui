package main

type FindNotesResult struct {
	Result []int  `json:"result"`
	Error  string `json:"error"`
}

type NotesInfoResult struct {
	Result []Note `json:"result"`
	Error  string `json:"error"`
}

type Note struct {
	NoteID    int      `json:"noteId"`
	Fields    Fields   `json:"fields"`
	Tags      []string `json:"tags"`
	ModelName string   `json:"modelName"`
}

// Fields represents the main fields for a Anki Note
type Fields map[string]interface{}
