package main

import (
	"flag"
	"fmt"
	"yadro-microservices/processing"
)

func main() {
	text := flag.String("s", "", "Input text to normalize")
	stopWordsFile := flag.String("f", "", "File containing stop words")
	language := flag.String("l", "en", "Language code for stemming and stop words")
	flag.Parse()

	tp := processing.NewTextProcessor(*text, *language)

	err := tp.RemoveStopWords(*stopWordsFile)
	if err != nil {
		fmt.Println("Error while removing stop words:", err)
		return
	}

	err = tp.Normalize()
	if err != nil {
		fmt.Println("Error while normalizing:", err)
		return
	}

	tp.RemoveDuplicates()

	fmt.Println(tp.GetText())
}
