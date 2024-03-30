package main

import (
	"flag"
	"fmt"
	"strings"
	"yadro-microservices/processing"
)

func main() {
	text := flag.String("s", "", "Input text to normalize")
	stopWordsFile := flag.String("f", "", "File containing stop words")
	language := flag.String("l", "en", "Language code for stemming and stop words")
	flag.Parse()

	tp := processing.NewTextProcessor(*language, *stopWordsFile)
	res, err := tp.FullProcess(*text)
	if err != nil {
		fmt.Println("Error while processing text:", err)
		return
	}

	fmt.Println(strings.Join(res, " "))
}
