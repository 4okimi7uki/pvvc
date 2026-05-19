package slack

type TextObject struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

type Block struct {
	Type     string       `json:"type"`
	Text     *TextObject  `json:"text,omitempty"`
	Elements []TextObject `json:"elements,omitempty"`
}

type blockPayload struct {
	Blocks []Block `json:"blocks"`
}
