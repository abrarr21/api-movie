package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type envelope map[string]interface{}

func (app *application) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("Invalid ID parameter")
	}

	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	jsonData = append(jsonData, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonData)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, destination interface{}) error {
	err := json.NewDecoder(r.Body).Decode(destination)
	if err != nil {
		// If theres an error during decoding, start the triage
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// Use the errors.As() function to check whether the error has the type *json.SyntaxError. If it does, then return a plain-english error message which includes the location of the problem.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("Body contains badly-formed JSON (at char %d)", syntaxError.Offset)

		// In some circumstances Decode() may also return an io.ErrUnexpectedEOF error for syntax errors in the JSON. So we check for this using errors.Is() and return a generic error message.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("Body contains badly-formed JSON")

			// Likewise, catch any *json.UnmarshalTypeError errors. These occur when the JSON value is the wrong type for the target destination. If the error relates to a specific field, then we include that in our error message to make it easier for the client to debug.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("Body contain incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("Body contains incorrect JSON type (at char %d)", unmarshalTypeError.Offset)

			// An io.EOF error will be returned by Decode() if the request body is empty. We check for this with errors.Is() and return a plain-english error message instead.
		case errors.Is(err, io.EOF):
			return errors.New("Body must not be empty")

			// A json.InvalidUnmarshalError error will be returned if we pass a non-nil pointer to Decode(). We catch this and panic, rather than returning an error to our handler. At the end of this chapter we'll talk about panicking versus returning errors, and discuss why it's an appropriate thing to do in this specific situation
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

			// For anything else, return the error message as-is.
		default:
			return err
		}
	}
	return nil
}
