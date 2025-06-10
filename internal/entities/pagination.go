package entities

type PaginationMetadata struct {
	CurrentPage  int
	PageSize     int
	FirstPage    int
	LastPage     int
	TotalRecords int
}

// GetPaginationMetadata generates pagination metadata based on the provided page,
// pageSize, and total record count.
//
// Parameters:
//   - page:      The current page number (starting from 1).
//   - pageSize:  The number of records per page.
//   - count:     The total number of records available.
//
// Returns:
//
//	PaginationMetadata struct filled with pagination details. If count or pageSize
//	is zero, returns an empty PaginationMetadata struct.
func GetPaginationMetadata(page, pageSize int, count int64) PaginationMetadata {
	if count == 0 || pageSize == 0 {
		return PaginationMetadata{}
	}

	return PaginationMetadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (int(count) + pageSize - 1) / pageSize,
		TotalRecords: int(count),
	}
}
