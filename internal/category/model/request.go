package model

type CreateCategoryRequest struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Status      int     `json:"status"`
}

type UpdateCategoryRequest struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Status      int     `json:"status"`
}
