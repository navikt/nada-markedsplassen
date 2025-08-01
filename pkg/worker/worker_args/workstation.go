package worker_args

import "riverqueue.com/riverpro"

const (
	WorkstationJobKind   = "workstation_job"
	WorkstationStartKind = "workstation_start"
	WorkstationStopKind  = "workstation_stop"

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

type WorkstationStop struct {
	Ident     string `json:"ident" river:"unique,sequence"`
	RequestID string `json:"request_id"`
}

func (WorkstationStop) Kind() string {
	return WorkstationStopKind
}

func (WorkstationStop) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:              true,
		ExcludeKind:         true,
		ContinueOnDiscarded: true,
	}
}

type WorkstationStart struct {
	Ident string `json:"ident" river:"unique,sequence"`
}

func (WorkstationStart) Kind() string {
	return WorkstationStartKind
}

func (WorkstationStart) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:              true,
		ExcludeKind:         true,
		ContinueOnDiscarded: true,
	}
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
		ByArgs:              true,
		ExcludeKind:         true,
		ContinueOnDiscarded: true,
	}
}

type WorkstationDisconnectJob struct {
	Ident string   `json:"ident" river:"sequence"`
	Hosts []string `json:"hosts"`
}

func (WorkstationDisconnectJob) Kind() string {
	return WorkstationDisconnectKind
}

func (WorkstationDisconnectJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:              true,
		ExcludeKind:         true,
		ContinueOnDiscarded: true,
	}
}

type WorkstationNotifyJob struct {
	Ident     string   `json:"ident" river:"sequence"`
	Hosts     []string `json:"hosts"`
	RequestID string   `json:"request_id"`
}

func (WorkstationNotifyJob) Kind() string {
	return WorkstationNotifyKind
}

func (WorkstationNotifyJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:              true,
		ExcludeKind:         true,
		ContinueOnDiscarded: true,
	}
}
