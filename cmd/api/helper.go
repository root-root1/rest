package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/root-root1/rest/internal/validator"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (app *Application) readJSON(w http.ResponseWriter, r *http.Request, dist interface{}) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dist)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("Body Containing Badly-formed JSON (At Character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("Body containing Badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("Body Contains Incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("Body Contains Incorrect JSON type (At Character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return fmt.Errorf("Body must not be Empty")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("Body Contains Unknown key %s", fieldName)

		case err.Error() == "http: request body to large":
			return fmt.Errorf("Body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("Body must contains single json Value")
	}
	return nil
}

func (app *Application) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		return 0, errors.New("Invalid Id Parameter")
	}

	return id, nil
}

type envelope map[string]interface{}

func (app *Application) writeJSON(w http.ResponseWriter, status int, data envelope, header http.Header) error {
	//js, err := json.MarshalIndent(data, "", "\t")
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range header {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *Application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *Application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *Application) readINT(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	val := qs.Get(key)

	if val == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(val)

	if err != nil {
		v.AddError(key, "must be an Integer Value")
		return defaultValue
	}

	return i
}
