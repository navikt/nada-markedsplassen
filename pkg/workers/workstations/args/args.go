package args

type WorkstationJob struct {
	Ident string `json:"ident" river:"unique"`

	Name  string `json:"name"`
	Email string `json:"email"`

	MachineType     string   `json:"machine_type"`
	ContainerImage  string   `json:"container_image"`
	URLAllowList    []string `json:"url_allow_list"`
	OnPremAllowList []string `json:"on_prem_allow_list"`
}

func (WorkstationJob) Kind() string {
	return "workstation"
}
