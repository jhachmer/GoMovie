# GoList
<div style="text-align:center"><img alt="gopherize.me" src="assets/gopher_small.png" /></div>

## About
GoList is a basic web app I build for personal use.
Someday it is supposed to be an app to manage a watchlist in the browser saving data regarding
who recommended the movie, if it is already watched and user comments in a SQLite database.\
Currently, there is only an info page accessible by a movies IMDb ID or Title and Year

### TODO:
  - Index Page
  - Search
  - Delete Entries

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
