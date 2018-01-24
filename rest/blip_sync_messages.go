package rest

import (
	"fmt"
	"math"
	"strings"

	"github.com/couchbase/go-blip"
	"github.com/couchbase/sync_gateway/base"
	"github.com/couchbase/sync_gateway/channels"
	"github.com/couchbase/sync_gateway/db"
)

const (
	BlipPropertySince      = "since"
	BlipPropertyBatch      = "batch"
	BlipPropertyContinuous = "continuous"
	BlipPropertyActiveOnly = "active_only"
	BlipPropertyFilter     = "filter"
	BlipPropertyChannels   = "channels"
)

// Function signature for something that generates a sequence id
type SequenceIDGenerator func() db.SequenceID

// Helper for handling BLIP subChanges requests.  Supports Stringer() interface to log aspects of the request.
type subChanges struct {
	rq                    *blip.Message       // The underlying BLIP message for this subChanges request
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
	return int(getRestrictedIntFromString(s.rq.Properties[BlipPropertyBatch], 200, 10, math.MaxUint64, true))
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

func (s *subChanges) String() string {

	channels, _ := s.channels()

	return fmt.Sprintf(
		"Since: %v Continuous: %v ActiveOnly: %v.  Filter: %v.  Channels: %v",
		s.since(),
		s.continuous(),
		s.activeOnly(),
		s.filter(),
		channels,
	)
}
