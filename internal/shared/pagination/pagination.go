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

	skip, skipErr := parseIntParam(query.Get("skip"))
	take, takeErr := parseIntParam(query.Get("take"))

	if skipErr != nil || takeErr != nil {
		return Params{}, ErrInvalid
	}

	return FromValues(skip, take)
}

// FromValues applies the pagination defaults and limits to optional typed values.
func FromValues(skip, take *int) (Params, error) {
	resolvedSkip := DefaultSkip
	if skip != nil {
		resolvedSkip = *skip
	}

	resolvedTake := DefaultTake
	if take != nil {
		resolvedTake = *take
	}

	if resolvedSkip < 0 || resolvedTake < 0 {
		return Params{}, ErrInvalid
	}

	if resolvedTake > MaxTake {
		resolvedTake = MaxTake
	}

	return Params{Skip: resolvedSkip, Take: resolvedTake}, nil
}

func parseIntParam(raw string) (*int, error) {
	if raw == "" {
		return nil, nil
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return nil, ErrInvalid
	}

	return &v, nil
}
