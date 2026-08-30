package pagination

import (
	"errors"
	"net/http"
	"strconv"
)

const (
	DefaultSkip = 0
	DefaultTake = 10
	MaxTake     = 100
)

var ErrInvalid = errors.New("invalid pagination params")

type Params struct {
	Skip int
	Take int
}

func Parse(r *http.Request) (Params, error) {
	query := r.URL.Query()

	skip, skipErr := parseIntParam(query.Get("skip"), DefaultSkip)
	take, takeErr := parseIntParam(query.Get("take"), DefaultTake)

	if skipErr != nil || takeErr != nil {
		return Params{}, ErrInvalid
	}

	if take > MaxTake {
		take = MaxTake
	}

	return Params{Skip: skip, Take: take}, nil
}

func parseIntParam(raw string, defaultValue int) (int, error) {
	if raw == "" {
		return defaultValue, nil
	}

	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return 0, ErrInvalid
	}

	return v, nil
}
