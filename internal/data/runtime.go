package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var InvalidRuntimeJsonFormat = errors.New("Invalid Runtime Format")

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d min", r)

	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}

func (r *Runtime) UnmarshalJSON(jsonvalue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonvalue))
	if err != nil {
		return InvalidRuntimeJsonFormat
	}

	part := strings.Split(unquotedJSONValue, " ")

	if len(part) != 2 || part[1] != "mins" {
		return InvalidRuntimeJsonFormat
	}

	i, err := strconv.ParseInt(part[0], 10, 32)
	if err != nil {
		return InvalidRuntimeJsonFormat
	}

	*r = Runtime(i)

	return nil
}
