package datastore

import (
	"context"
)

// EventWriterCounter is used to write VisitEvent entries, and count how many
// events match the supplied mask.
type EventWriterCounter interface {
	Write(context.Context, *VisitEvent) error
	Count(context.Context, *QueryEvent) (int, error)
}
