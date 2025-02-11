package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jhachmer/gomovie/internal/parser"
)

func main() {
	fmt.Println("Parsing...")
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "invalid number of args (%d)\n", len(os.Args))
		os.Exit(1)
	}
	args, err := parseArgs()
	if err != nil {
		printUsage()
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	err = parser.ParseCSV(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func parseArgs() (*parser.ParseArgs, error) {
	if len(os.Args) < 6 {
		return nil, fmt.Errorf("insufficient arguments: expected 5, got %d", len(os.Args)-1)
	}

	var args parser.ParseArgs
	args.Path = os.Args[1]

	indices := []struct {
		arg   string
		dest  *int
		label string
	}{
		{os.Args[2], &args.TitleIndex, "title index"},
		{os.Args[3], &args.YearIndex, "year index"},
		{os.Args[4], &args.WatchedIndex, "watched index"},
		{os.Args[5], &args.RecommenderIndex, "recommender index"},
	}

	for _, index := range indices {
		value, err := strconv.Atoi(index.arg)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %v", index.label, err)
		}
		*index.dest = value
	}

	return &args, nil
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gomovie_parser <path> <titleIndex> <yearIndex> <watchedIndex> <recommenderIndex>")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <path>             The path to the input file.")
	fmt.Println("  <titleIndex>       The index of the title column in the input.")
	fmt.Println("  <yearIndex>        The index of the year column in the input.")
	fmt.Println("  <watchedIndex>     The index of the watched column in the input.")
	fmt.Println("  <recommenderIndex> The index of the recommender column in the input.")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  gomovie_parser data.csv 0 1 2 3")
}
