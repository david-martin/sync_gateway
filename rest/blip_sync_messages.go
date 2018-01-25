package rest

import (
	"fmt"
	"math"
	"strings"

	"bytes"
	"github.com/couchbase/go-blip"
	"github.com/couchbase/sync_gateway/base"
	"github.com/couchbase/sync_gateway/channels"
	"github.com/couchbase/sync_gateway/db"
)

// Function signature for something that generates a sequence id
type SequenceIDGenerator func() db.SequenceID

// Helper for handling BLIP subChanges requests.  Supports Stringer() interface to log aspects of the request.
type subChanges struct {
	rq                    *blip.Message       // The underlying BLIP message
	logger                base.SGLogger       // A logger object which might encompass more state (eg, blipContext id)
	sinceZeroValueCreator SequenceIDGenerator // A sequence generator for creating zero'd since values
}

// Create a new subChanges helper
func newSubChanges(rq *blip.Message, logger base.SGLogger, sinceZeroValueCreator SequenceIDGenerator) *subChanges {
	return &subChanges{
		rq:                    rq,
		logger:                logger,
		sinceZeroValueCreator: sinceZeroValueCreator,
	}
}

func (s *subChanges) since() db.SequenceID {

	// Depending on the db sequence type, use correct zero sequence for since value
	sinceSequenceId := s.sinceZeroValueCreator()

	if sinceStr, found := s.rq.Properties[BlipPropertySince]; found {
		var err error
		if sinceSequenceId, err = db.ParseSequenceIDFromJSON([]byte(sinceStr)); err != nil {
			s.logger.LogTo("Sync", "%s: Invalid sequence ID in 'since': %s", s.rq, sinceStr)
			sinceSequenceId = db.SequenceID{}
		}
	}

	return sinceSequenceId

}

func (s *subChanges) batchSize() int {
	return int(getRestrictedIntFromString(s.rq.Properties[BlipPropertyBatch], BlipDefaultBatchSize, 10, math.MaxUint64, true))
}

func (s *subChanges) continuous() bool {
	continuous := false
	if val, found := s.rq.Properties[BlipPropertyContinuous]; found && val != "false" {
		continuous = true
	}
	return continuous
}

func (s *subChanges) activeOnly() bool {
	return (s.rq.Properties[BlipPropertyActiveOnly] == "true")
}

func (s *subChanges) filter() string {
	return s.rq.Properties[BlipPropertyFilter]
}

func (s *subChanges) channels() (channels string, found bool) {
	channels, found = s.rq.Properties[BlipPropertyChannels]
	return channels, found
}

func (s *subChanges) channelsExpandedSet() (resultChannels base.Set, err error) {
	channelsParam, found := s.rq.Properties[BlipPropertyChannels]
	if !found {
		return nil, fmt.Errorf("Missing 'channels' filter parameter")
	}
	channelsArray := strings.Split(channelsParam, ",")
	return channels.SetFromArray(channelsArray, channels.ExpandStar)
}

// Satisfy fmt.Stringer interface for dumping attributes of this subChanges request to logs
func (s *subChanges) String() string {

	buffer := bytes.NewBufferString("")

	buffer.WriteString(fmt.Sprintf("Since:%v ", s.since()))

	continuous := s.continuous()
	if continuous {
		buffer.WriteString(fmt.Sprintf("Continuous:%v ", continuous))
	}

	activeOnly := s.activeOnly()
	if activeOnly {
		buffer.WriteString(fmt.Sprintf("ActiveOnly:%v ", activeOnly))
	}

	filter := s.filter()
	if len(filter) > 0 {
		buffer.WriteString(fmt.Sprintf("Filter:%v ", filter))
		channels, found := s.channels()
		if found {
			buffer.WriteString(fmt.Sprintf("Channels:%v ", channels))
		}
	}

	batchSize := s.batchSize()
	if batchSize != int(BlipDefaultBatchSize) {
		buffer.WriteString(fmt.Sprintf("BatchSize:%v ", s.batchSize()))
	}

	return buffer.String()

}

type setCheckpoint struct {
	rq *blip.Message // The underlying BLIP message
}

func newSetCheckpoint(rq *blip.Message) *setCheckpoint {
	return &setCheckpoint{
		rq: rq,
	}
}

func (s *setCheckpoint) client() string {
	return s.rq.Properties[BlipPropertyClient]
}

func (s *setCheckpoint) rev() string {
	return s.rq.Properties[BlipPropertyRev]
}

func (s *setCheckpoint) String() string {

	buffer := bytes.NewBufferString("")

	buffer.WriteString(fmt.Sprintf("Client:%v ", s.client()))

	rev := s.rev()
	if len(rev) > 0 {
		buffer.WriteString(fmt.Sprintf("Rev:%v ", rev))
	}

	return buffer.String()

}

type addRevision struct {
	rq *blip.Message // The underlying BLIP message
}

func newAddRevision(rq *blip.Message) *addRevision {
	return &addRevision{
		rq: rq,
	}
}

func (a *addRevision) id() (id string, found bool) {
	id, found = a.rq.Properties[BlipPropertyId]
	return id, found
}

func (a *addRevision) rev() (rev string, found bool) {
	rev, found = a.rq.Properties[BlipPropertyRev]
	return rev, found
}

func (a *addRevision) deleted() (deleted string, found bool) {
	deleted, found = a.rq.Properties[BlipPropertyDeleted]
	return deleted, found
}

func (a *addRevision) sequence() (sequence string, found bool) {
	sequence, found = a.rq.Properties[BlipPropertySequence]
	return sequence, found
}

func (a *addRevision) String() string {

	buffer := bytes.NewBufferString("")

	if id, foundId := a.id(); foundId {
		buffer.WriteString(fmt.Sprintf("Id:%v ", id))
	}

	if rev, foundRev := a.rev(); foundRev {
		buffer.WriteString(fmt.Sprintf("Rev:%v ", rev))
	}

	if deleted, foundDeleted := a.deleted(); foundDeleted {
		buffer.WriteString(fmt.Sprintf("Deleted:%v ", deleted))
	}

	if sequence, foundSequence := a.sequence(); foundSequence == true {
		buffer.WriteString(fmt.Sprintf("Sequence:%v ", sequence))
	}

	return buffer.String()

}

type getAttachment struct {
	rq *blip.Message // The underlying BLIP message
}

func newGetAttachment(rq *blip.Message) *getAttachment {
	return &getAttachment{
		rq: rq,
	}
}


func (g *getAttachment) digest() string {
	return g.rq.Properties[BlipPropertyDigest]
}


func (g *getAttachment) String() string {

	buffer := bytes.NewBufferString("")

	buffer.WriteString(fmt.Sprintf("Digest:%v ", g.digest()))

	return buffer.String()

}