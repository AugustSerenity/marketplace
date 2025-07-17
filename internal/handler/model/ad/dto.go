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
