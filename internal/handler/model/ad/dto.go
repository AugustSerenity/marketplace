package ad

type CreateRequest struct {
	Title       string  `json:"title" validate:"required,max=100"`
	Description string  `json:"description" validate:"required,max=1000"`
	ImageURL    string  `json:"image_url" validate:"required,url"`
	Price       float64 `json:"price" validate:"required,gt=0"`
}

type Response struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
	AuthorID    int64   `json:"author_id"`
}

type ListRequest struct {
	Page      int     `json:"page" validate:"gte=1"`
	PageSize  int     `json:"page_size" validate:"gte=1,lte=100"`
	SortBy    string  `json:"sort_by" validate:"oneof=created_at price"`
	SortOrder string  `json:"sort_order" validate:"oneof=asc desc"`
	MinPrice  float64 `json:"min_price" validate:"gte=0"`
	MaxPrice  float64 `json:"max_price" validate:"gte=0"`
}

type ListResponse struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
	AuthorLogin string  `json:"author_login"`
	IsOwner     bool    `json:"is_owner"`
}
