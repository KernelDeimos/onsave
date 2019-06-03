package inputfilters

import (
	"errors"
	"path/filepath"
)

type PathFiltersI interface {
	NormaliseToRel(path string) (string, error)
}

type PathFilters struct {
	WorkingDirectory string
}

func (filters PathFilters) NormaliseToRel(path string) (string, error) {
	path = filepath.Join(filters.WorkingDirectory, path)
	npath, err := filepath.Rel(filters.WorkingDirectory, path)
	if err != nil {
		return npath, errors.New("could not normalise path: " + err.Error())
	}
	return npath, nil
}
