package data

import (
	"github.com/root-root1/rest/internal/validator"
	"math"
	"strings"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

type Metadata struct {
	CurrentPage int `json:"current_page,omitempty"`
	PageSize    int `json:"page_size"`
	FirstPage   int `json:"first_page"`
	LastPage    int `json:"last_page"`
	TotalRecord int `json:"total_record"`
}

func CalculateMetadata(totalRecord int, page int, pageSize int) Metadata {
	if totalRecord == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage: page,
		PageSize:    pageSize,
		FirstPage:   1,
		LastPage:    int(math.Ceil(float64(totalRecord) / float64(pageSize))),
		TotalRecord: totalRecord,
	}
}
func ValidateFilter(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "Page must be greater than 0")
	v.Check(f.Page <= 10_000_000, "page", "Page must be Less than 10 million")
	v.Check(f.PageSize > 0, "page_size", "Page Size must be Greater than 0")
	v.Check(f.PageSize <= 100, "page_size", "Page Size must be Less than 100 or Equal")

	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "Invalid sort Value")
}

func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("Unsafe sort parameter: " + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}
