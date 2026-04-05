package models

const (
	OrderAsc  = "asc"
	OrderDesc = "desc"
)

// WeighIn represents a single weigh-in record.
type WeighIn struct {
	ID        int64   `json:"id"`
	Weight    float64 `json:"weight"`
	CreatedAt string  `json:"created_at"`
	Source    string  `json:"source"`
	Notes     string  `json:"notes"`
	UpdatedAt string  `json:"updated_at"`
	DeletedAt string  `json:"deleted_at,omitempty"`
}

// ListOpts holds filtering/sorting options for listing weigh-ins.
type ListOpts struct {
	Since  string
	Until  string
	Source string
	Order  string // "asc" or "desc"
	Limit  int
}

// DeleteResult represents the output of a delete operation.
type DeleteResult struct {
	ID      int64 `json:"id"`
	Deleted bool  `json:"deleted"`
}
