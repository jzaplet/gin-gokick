package widget

import (
	"net/netip"
	"time"

	"github.com/google/uuid"

	"fixture/api"
	"fixture/owner"
)

//tsgen:assets/app/Widget/types/Widget.ts Widget
type widget struct {
	ID uuid.UUID `json:"id"`

	Count int `json:"count"`

	Ratio float64 `json:"ratio"`

	Active bool `json:"active"`

	Tags []string `json:"tags"`

	Note *string `json:"note"`

	Aliases *[]string `json:"aliases"`

	Hint string `json:"hint,omitempty"`

	Owner owner.Owner `json:"owner"`

	Deputy *owner.Owner `json:"deputy"`

	Members []owner.Owner `json:"members"`

	Status api.Key `json:"status"`

	Meta map[string]any `json:"meta"`

	CreatedAt time.Time `json:"createdAt"`

	SeenFrom *netip.Addr `json:"seenFrom"`

	Kind string `json:"type"`

	Secret string `json:"-"`

	internal string
}

var _ = widget{internal: ""}
