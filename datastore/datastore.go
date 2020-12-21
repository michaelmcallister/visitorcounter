package datastore

// EventWriterCounter is used to write VisitEvent entries, and count how many
// events match the supplied mask.
type EventWriterCounter interface {
	Write(*VisitEvent) error
	Count(*QueryEvent) (int, error)
}
