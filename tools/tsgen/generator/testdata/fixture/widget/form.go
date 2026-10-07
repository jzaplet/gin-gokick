package widget

import "fixture/owner"

//tsgen:assets/app/Widget/types/WidgetForm.ts WidgetForm noguard
type WidgetForm struct {
	Name string `json:"name"`

	Owner owner.Owner `json:"owner"`

	Note string `json:"note,omitempty"`
}
