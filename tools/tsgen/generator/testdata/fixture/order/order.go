package order

import "fixture/api"

//tsgen:assets/app/Order/types/Address.ts Address noguard
type Address struct {
	Street string `json:"street" binding:"required"`

	Kind string `json:"type" binding:"oneof=home work"`
}

//tsgen:assets/app/Order/types/Item.ts Item noguard
type Item struct {
	Name string `json:"name" binding:"required"`

	Tags []string `json:"tags" binding:"dive,min=1"`
}

//tsgen:assets/app/Order/types/OrderRequest.ts OrderRequest request
type OrderRequest struct {
	Billing Address `json:"billing"`

	Shipping *Address `json:"shipping"`

	Items []Item `json:"items" binding:"required,min=1,dive"`

	Labels []api.Key `json:"labels"`

	Note string `json:"note,omitempty"`
}
