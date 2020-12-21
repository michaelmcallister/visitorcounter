package datastore

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc32"
	"net"
	"time"

	bolt "go.etcd.io/bbolt"
)

const buckets = 10

// Bolt satisfies the EventWriterCounter interface for persisting visit requests.
type Bolt struct {
	c *bolt.DB
}

// VisitEvent is an entry that represents a single web request.
type VisitEvent struct {
	Time   time.Time
	Domain string
	IP     net.IP
}

// QueryEvent is used to search and count how many requests match the supplied
// query.
type QueryEvent struct {
	Domain string
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func hash(s string) uint32 {
	return crc32.ChecksumIEEE([]byte(s)) % buckets
}

// NewBolt creates and opens a database at the given path. If the file does not
// exist then it will be created automatically. Passing in nil options
// will cause Bolt to open the database with the default options.
func NewBolt(file string, o *bolt.Options) (*Bolt, error) {
	db, err := bolt.Open(file, 0666, o)
	if err != nil {
		return nil, err
	}
	return &Bolt{db}, nil
}

// Write persists the supplied VisitEvent requested, returning nil if successful
// otherwise an error.
func (b *Bolt) Write(e *VisitEvent) error {
	// Marshal VisitEvent into bytes.
	buf, err := json.Marshal(e)
	if err != nil {
		return err
	}

	bucket := hash(e.Domain)

	return b.c.Update(func(tx *bolt.Tx) error {
		bk, _ := tx.CreateBucketIfNotExists(itob(uint64(bucket)))
		id, _ := bk.NextSequence()
		return bk.Put(itob(id), buf)
	})
}

// Count will return the number of records that match the supplied QueryEvent.
// A negative integer will be returned in the case of an error.
func (b *Bolt) Count(q *QueryEvent) (int, error) {
	if q.Domain == "" {
		return 0, errors.New("empty domain")
	}

	bucket := hash(q.Domain)
	cnt := 0
	err := b.c.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(itob(uint64(bucket)))
		// We've not seen this domain before?
		if bkt == nil {
			return nil
		}
		return bkt.ForEach(func(_, v []byte) error {
			ev := &VisitEvent{}
			if err := json.Unmarshal(v, ev); err != nil {
				return err
			}
			if ev.Domain == q.Domain {
				cnt++
			}
			return nil
		})
	})
	if err != nil {
		return 0, err
	}
	return cnt, nil
}
