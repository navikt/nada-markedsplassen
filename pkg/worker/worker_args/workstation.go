package worker_args

import "riverqueue.com/riverpro"

const (
	WorkstationJobKind              = "workstation_job"
	WorkstationStartKind            = "workstation_start"
	WorkstationZonalTagBindingsKind = "workstation_zonal_tag_bindings"

	WorkstationConnectKind    = "workstation_connect"
	WorkstationDisconnectKind = "workstation_disconnect"
	WorkstationNotifyKind     = "workstation_notify"

	WorkstationQueue             = "workstation"
	WorkstationConnectivityQueue = "workstation_connectivity"
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

type WorkstationConnectJob struct {
	Ident string `json:"ident" river:"sequence"`
	Host  string `json:"host"`
}

func (WorkstationConnectJob) Kind() string {
	return WorkstationConnectKind
}

func (WorkstationConnectJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ExcludeKind: true,
	}
}

type WorkstationDisconnectJob struct {
	Ident string   `json:"ident" river:"sequence"`
	Hosts []string `json:"host"`
}

func (WorkstationDisconnectJob) Kind() string {
	return WorkstationDisconnectKind
}

func (WorkstationDisconnectJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ExcludeKind: true,
	}
}

type WorkstationNotifyJob struct {
	Ident     string   `json:"ident" river:"sequence"`
	Hosts     []string `json:"host"`
	RequestID string   `json:"request_id"`
}

func (WorkstationNotifyJob) Kind() string {
	return WorkstationNotifyKind
}

func (WorkstationNotifyJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ExcludeKind: true,
	}
}
