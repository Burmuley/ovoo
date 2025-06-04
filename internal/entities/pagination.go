package entities

type PaginationMetadata struct {
	CurrentPage  int
	PageSize     int
	FirstPage    int
	LastPage     int
	TotalRecords int
}

func GetPaginationMetadata(page, pageSize int, count int64) PaginationMetadata {
	if count == 0 {
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
