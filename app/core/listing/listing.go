package listing

//tsgen:assets/shared/Grid/types/SortDirection.ts SortDirection union
type Direction string

const Ascending Direction = "asc"

const Descending Direction = "desc"

const DefaultPerPage = 25

type Grid[S ~string] struct {
	Sorts []S

	Direction Direction

	PerPage int
}

type Query[S ~string] struct {
	Page int

	PerPage int

	Sort S

	Direction Direction
}

func (q Query[S]) Offset() int64 {
	return int64(q.Page-1) * int64(q.PerPage)
}

func (q Query[S]) Limit() int64 {
	return int64(q.PerPage)
}

type Result[T any] struct {
	Items []T `json:"items"`

	Total int64 `json:"total"`
}

type SelectableResult[T any] struct {
	Items []T `json:"items"`

	Total int64 `json:"total"`

	Selectable int64 `json:"selectable"`
}
