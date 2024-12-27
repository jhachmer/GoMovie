# GoList
![gopher](assets/gopher_small.png "https://gopherize.me/")

## About
GoList is a basic web app I build for personal use.
Someday it is supposed to be an app to manage a watchlist in the browser saving data regarding
who recommended the movie, if it is already watched and user comments in a SQLite database.\
Currently, there is only an info page accessible by a movies IMDb ID or Title and Year

### To build yourself you'll need:
  - Go (developed with 1.23)
    -  [go-task](https://taskfile.dev/) if you want to use Taskfile to build
  - [OMDB](https://www.omdbapi.com/) API key

### TODO:
  - Index Page
  - Search Bar
    - search for genres, year, already watched etc.
  - Delete Entries
  - ~~Split Genres and Actors~~
    - ~~separate db tables for them~~
  - Update Button for new movie info (poster, ratings)
  - Color Grading of rows

### Routes:
- GET /health : returns healthy if server is running
- GET /films/{imdb} : returns info page for movie with imdb id
- POST /films/{imdb} : posts a new entry for movie
- GET /films/{title}/{year} : returns info page for movie with title and year
- GET /{$} : index page for root route (really basic atm)

## Setup
Use either:
- Taskfile
  - you will need [go-task](https://taskfile.dev/) for this
  ```shell
    # be sure to provide an OMDb API key
    # either as environment variable or via .env file
    # cd into project root
    cd goList
    # run task with default build target (no args) to setup, test and build the application
    task
    # you can also run every task independently (test, clean, ...)
    # - task test
    # - task clean
    # execute file
    ./bin/golist_svr
  ```
- Docker
  ```shell
    # Run Docker client
    # cd into project root
    cd goList
    # build docker image e.g.
    docker build --tag docker-golist .
    # spin up the container
    # remember to pass your omdb api key as a env variable
    docker run -d --publish 8080:8080 -e OMDB_KEY=your_key docker-golist
  ```
