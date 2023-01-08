package data

import "github.com/root-root1/rest/internal/validator"

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func ValidateFilter(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "Page must be greater than 0")
	v.Check(f.Page <= 10_000_000, "page", "Page must be Less than 10 million")
	v.Check(f.PageSize > 0, "page_size", "Page Size must be Greater than 0")
	v.Check(f.PageSize <= 100, "page_size", "Page Size must be Less than 100 or Equal")

	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "Invalid sort Value")
}
