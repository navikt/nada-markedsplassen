package worker_args

const (
	WorkstationJobKind              = "workstation_job"
	WorkstationStartKind            = "workstation_start"
	WorkstationZonalTagBindingsKind = "workstation_zonal_tag_bindings"

	WorkstationQueue = "workstation"
)

type WorkstationJob struct {
	Ident string `json:"ident" river:"unique"`

	Name  string `json:"name"`
	Email string `json:"email"`

	MachineType    string `json:"machine_type"`
	ContainerImage string `json:"container_image"`
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

type WorkstationZonalTagBindingsJob struct {
	Ident     string   `json:"ident" river:"unique"`
	RequestID string   `json:"request_id"`
	Hosts     []string `json:"hosts"`
}

func (WorkstationZonalTagBindingsJob) Kind() string {
	return WorkstationZonalTagBindingsKind
}
