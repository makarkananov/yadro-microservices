package core

// Comic includes the URL of the comic image and a list of keywords associated with the comic.
type Comic struct {
	Img      string   `json:"url"`
	Keywords []string `json:"keywords"`
}
