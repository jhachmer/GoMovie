# GoList
![gopher](assets/gopher_small.png "https://gopherize.me/")

## About
GoList is a basic web app I build for personal use.
Someday it is supposed to be an app to manage a watchlist in the browser, saving data regarding
who recommended the movie, if we already watched it and comments about the movie in a SQLite database.\

### Example Overview
![overview](assets/overview.png)
### Example Info Page
![info](assets/info.png)




### To build yourself you'll need:
  - Go (developed with 1.23)
    -  [go-task](https://taskfile.dev/) if you want to use Taskfile to build
  - [OMDB](https://www.omdbapi.com/) API key

### TODO:
  - ~~Index Page~~
  - ~~Search Bar~~
    - search for genres, year, already watched etc.
    - kinda done, but needs additional work
  - ~~Delete Entries~~
  - ~~Split Genres and Actors~~
    - ~~separate db tables for them~~
  - Update Button for new movie info (poster, ratings)
    - especially recently announed movies have a placeholder image as poster and obviously no ratings, updating them should provide the ratings and poster at the time of updating
  - ~~Color Grading of rows, to show if they are watched~~
  - checkbox to toggle showing only unwatched movies on overview

### Routes:
- GET /health : returns healthy if server is running
- GET /login : displays login page
- POST /login : checks credentials provided by form values of username and password
- GET /overview : displays overview page of all movies in database
- GET /films/{imdb} : returns info page for movie with imdb id
- POST /films/{imdb} : posts a new entry for movie
- PUT /films/{imdb} : changes the entry saved for that movie
- DELETE /films/{imdb} : deletes entry belonging to movie (does not delete movie from db, maybe later)
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
