package service

import "time"

type JobHeader struct {
	ID        int64      `json:"id"`
	StartTime time.Time  `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
	State     JobState   `json:"state"`
	Duplicate bool       `json:"duplicate"`
	Errors    []string   `json:"errors"`
}

type JobState string

const (
	JobStateCompleted JobState = "COMPLETED"
	JobStateRunning   JobState = "RUNNING"
	JobStateFailed    JobState = "FAILED"
	JobStatePending   JobState = "PENDING"
)
