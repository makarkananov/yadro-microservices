package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"yadro-microservices/internal/core"
	"yadro-microservices/pkg/database"
	"yadro-microservices/pkg/words"
	"yadro-microservices/pkg/xkcd"
)

func main() {
	// Parse command line flags
	configPath := flag.String("c", "config.yaml", "Path to configuration file")
	outputFlag := flag.Bool("o", false, "Flag to output data to screen")
	maxComicsShown := flag.Int("n", 10, "Flag to limit the number of comics")
	flag.Parse()

	// Initialize and load configuration from file
	viper.SetConfigFile(*configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Panic("Error loading configuration:", err)
	}

	xkcdClient := xkcd.NewClient(viper.GetString("source_url"))

	// Retrieve comics data from XKCD API
	maxComics := viper.GetInt("max_comics_load")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a channel to handle OS signals for program termination
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Wait for a termination signal
	go func() {
		<-sigCh
		fmt.Println("Received cancellation. Graceful shutdown...")
		cancel()
	}()

	fmt.Println("Retrieving comics data from xkcd.com...")
	comicsResponses, err := xkcdClient.GetComics(ctx, maxComics)
	if err != nil {
		fmt.Println("Error retrieving some comics data:", err)
	}

	// Convert XKCD comics data to internal representation using TextProcessor
	tp := words.NewTextProcessor("en", "")
	comicsMap := make(map[int]*core.Comic, len(comicsResponses))
	for i := range comicsResponses {
		comicText := strings.Join([]string{comicsResponses[i].Alt +
			comicsResponses[i].Transcript +
			comicsResponses[i].Title}, " ")

		// Extract keywords from the comic text
		kw, err := tp.FullProcess(comicText)
		if err != nil {
			log.Panic("Error extracting keywords:", err)
		}

		comicsMap[comicsResponses[i].Num] = &core.Comic{
			Img:      comicsResponses[i].Img,
			Keywords: kw,
		}
	}

	fmt.Println("Saving comics data to JSON database...")

	// Initialize JSON database
	db := database.NewJSONDB(viper.GetString("db_file"))

	// Save comics data to JSON database
	err = db.SaveJSON(comicsMap)
	if err != nil {
		log.Panic("Error saving comics data to JSON database:", err)
	}

	if *outputFlag {
		// Output comics data to screen
		t := table.NewWriter()

		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"#", "Img URL", "Keywords"})

		// Sort keys of comicsMap
		var sortedKeys []int
		for num := range comicsMap {
			sortedKeys = append(sortedKeys, num)
		}
		sort.Ints(sortedKeys)

		// Iterate over sorted keys
		for _, num := range sortedKeys {
			t.AppendRow([]interface{}{
				num,
				comicsMap[num].Img,
				text.WrapText(strings.Join(comicsMap[num].Keywords, " "), 70),
			})
			t.AppendSeparator()

			// Break the loop if maxComicsShown is reached
			if t.Length() >= *maxComicsShown {
				break
			}
		}

		t.SetStyle(table.StyleLight)
		t.Render()
	}
}
