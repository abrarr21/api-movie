package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/abrarr21/api-movie/internals/data"
	"github.com/abrarr21/api-movie/internals/validator"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err) //helper function
		return
	}

	//Copy the values from the input struct to a new Movie struct
	Movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	//an instance of the Validator struct
	v := validator.New()

	if data.ValidateMovie(v, Movie); !v.Valid() {
		app.faliedValidationResponse(w, r, v.Errors)
		return
	}

	//Dumping the content of the input struct in a http response
	fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r) //helper function
		return
	}

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Avengers",
		Runtime:   102,
		Genres:    []string{"Sci-Fi", "war", "Adventure"},
		Version:   1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
