package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/jhachmer/gomovie/internal/api"
	"github.com/jhachmer/gomovie/internal/util"
)

const IMDbIDPattern = `^tt\d{7,8}$`

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[gomovie-search] ")
	var imdbID string
	var title string
	var year string
	var searchType string

	flag.StringVar(&imdbID, "id", "", "imdb id")
	flag.StringVar(&title, "title", "", "movie title")
	flag.StringVar(&year, "year", "", "release year")
	flag.StringVar(&searchType, "type", "", "search type")
	flag.Parse()

	if imdbID != "" {
		if !regexp.MustCompile(IMDbIDPattern).MatchString(imdbID) {
			log.Fatalf("id %s is not a valid id", imdbID)
		}
		mov, err := api.MovieFromID(imdbID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s (%s) [%s]\n", mov.Title, mov.Year, mov.ImdbID)
		fmt.Printf("  Director: %s\n", mov.Director)
		fmt.Printf("  Writer: %s\n", mov.Writer)
		fmt.Printf("  Actors: %s\n", mov.Actors)
		fmt.Printf("  Genre: %s\n", mov.Genre)
		fmt.Printf("  Runtime: %s\n", mov.Runtime)
		fmt.Printf("  Country: %s\n", mov.Country)
		fmt.Printf("  Rating (Votes): %s (%s)\n", mov.ImdbRating, mov.ImdbVotes)
		fmt.Printf("  Plot: %s\n", mov.Plot)
		os.Exit(0)
	} else if title != "" {
		searchQuery := api.SearchQueryRequest{
			Title: title,
			Year:  year,
			Type:  searchType,
		}
		fmt.Printf("Search Query: %v\n", searchQuery)
		result, err := api.QueryOMDb(searchQuery)
		if err != nil {
			log.Fatal(err)
		}
		jsonString, err := util.JSONString(result)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(jsonString)
		os.Exit(0)
	}
	log.Println("please provide either an IMDb ID or a title to search for")
	os.Exit(1)
}
