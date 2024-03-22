package core

import (
	"fmt"
	"strings"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

func ParseJpSentence(input string) string {
	tagger, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	// change all spaces to full-width spaces of input
	// remove all spaces
	// input = strings.Replace(input, " ", "　", -1)
	input = strings.Replace(input, " ", "", -1)
	input = strings.Replace(input, "　", "", -1)

	seg := tagger.Tokenize(input)
	var rawSentence string
	for _, token := range seg {
		reading, hasReading := token.Reading()
		if hasReading {
			rawSentence += reading + "　"
		} else {
			rawSentence += fmt.Sprintf("%s　", token.Surface)
		}
	}

	return rawSentence
}
