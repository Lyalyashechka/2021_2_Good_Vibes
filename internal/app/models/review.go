package models

type Review struct {
	UserId    int    `json:"-"`
	UserName  string `json:"username,omitempty"`
	ProductId int    `json:"product_id,omitempty" validate:"required"`
	Rating    int    `json:"rating" validate:"required"`
	Text      string `json:"text" validate:"required"`
}
