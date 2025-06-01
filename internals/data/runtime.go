package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Unmarshaler shows this error if we're unable to parse the json string.
var ErrInvalidRuntimeFormat = errors.New("Invalid Runtime format")

// Declare a custom Runtime type, which has the underlying type int32 (the same as our Movie struct field).
type Runtime int32

// Implement a MarshalJSON() method on the Runtime type so that it satisfies the json.Marshaler interface. This should return the JSON-encoded value for the movie runtime (in our case, it will return a string in the format "<runtime> mins").
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d minutes", r)

	quotedJSONvalue := strconv.Quote(jsonValue)

	return []byte(quotedJSONvalue), nil
}

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	unqoutedJsonValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	parts := strings.Split(unqoutedJsonValue, " ")

	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = Runtime(i)

	return nil
}
