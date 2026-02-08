package model

import "time"

type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	// V1: StatusFailed Status = "failed"
)

type Priority string

const (
	P0 Priority = "P0"
	P1 Priority = "P1"
	P2 Priority = "P2"
)

type PushPolicy string

const (
	PushPolicyPush        PushPolicy = "push"
	PushPolicyArchiveOnly PushPolicy = "archive_only"
)

type Entry struct {
	ID          int64
	SourceID    string
	Title       string
	URL         string
	URLRaw      string
	PublishedAt *time.Time
	FetchedAt   time.Time
	Fingerprint string

	Status     Status
	Priority   Priority
	PushPolicy PushPolicy

	Tags    []string
	Summary string

	PayloadJSON string
}
