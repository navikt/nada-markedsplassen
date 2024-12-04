package worker_args

const (
	WorkstationJobKind             = "workstation_job"
	WorkstationStartKind           = "workstation_start"
	WorkstationZonalTagBindingKind = "workstation_zonal_tag_binding"

	WorkstationQueue = "workstation"
)

type WorkstationJob struct {
	Ident string `json:"ident" river:"unique"`

	Name  string `json:"name"`
	Email string `json:"email"`

	MachineType               string   `json:"machine_type"`
	ContainerImage            string   `json:"container_image"`
	URLAllowList              []string `json:"url_allow_list"`
	OnPremAllowList           []string `json:"on_prem_allow_list"`
	DisableGlobalURLAllowList bool     `json:"disable_global_url_allow_list"`
}

func (WorkstationJob) Kind() string {
	return WorkstationJobKind
}

type WorkstationStart struct {
	Ident string `json:"ident" river:"unique"`
}

func (WorkstationStart) Kind() string {
	return WorkstationStartKind
}

type WorkstationZonalTagBindingJob struct {
	Ident             string `json:"ident"`
	Action            string `json:"action"`
	Zone              string `json:"zone"`
	Parent            string `json:"parent"`
	TagValue          string `json:"tagValue"`
	TagNamespacedName string `json:"tagNamespacedName"`
}

func (WorkstationZonalTagBindingJob) Kind() string {
	return WorkstationZonalTagBindingKind
}
