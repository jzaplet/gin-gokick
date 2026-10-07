package owner

//tsgen:assets/app/Owner/types/Owner.ts Owner
type Owner struct {
	Name string `json:"name"`

	Peers []Owner `json:"peers"`
}
