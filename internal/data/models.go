package data

import (
	"database/sql"
	"errors"
)

var (
	ErrorRecordNotFound = errors.New("Record Not found")
	ErrEditConflict     = errors.New("Edit Conflict")
)

type Models struct {
	Movies interface {
		Insert(movie *Movie) error
		Get(id int64) (*Movie, error)
		Update(movie *Movie) error
		Delete(id int64) error
	}
}

func NewModel(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{db: db},
	}
}

func NewMockModel() Models {
	return Models{
		Movies: MockMovieModel{},
	}
}
