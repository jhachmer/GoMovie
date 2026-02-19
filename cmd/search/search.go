package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"

	"github.com/jhachmer/gomovie/internal/types"
)

const IMDbIDPattern = `^tt\d{7,8}$`

func main() {
	var imdbID string
	flag.StringVar(&imdbID, "id", "", "imdb id")

	flag.Parse()
	if imdbID == "" {
		log.Fatal("-id flag is required")
	}

	if !regexp.MustCompile(IMDbIDPattern).MatchString(imdbID) {
		log.Fatalf("id %s is not a valid id", imdbID)
	}

	mov, err := types.MovieFromID(imdbID)
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
}
