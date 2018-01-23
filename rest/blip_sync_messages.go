package rest

import (
	"fmt"
	"github.com/couchbase/go-blip"
	"github.com/couchbase/sync_gateway/base"
	"github.com/couchbase/sync_gateway/db"
	"math"
	"strings"
	"github.com/couchbase/sync_gateway/channels"
)

const (
	BlipPropertySince      = "since"
	BlipPropertyBatch      = "batch"
	BlipPropertyContinuous = "continuous"
	BlipPropertyActiveOnly = "active_only"
	BlipPropertyFilter     = "filter"
	BlipPropertyChannels   = "channels"
)

type subChanges struct {
	rq              *blip.Message
	logger          base.SGLogger
	sinceSequenceId db.SequenceID
}

func newSubChanges(rq *blip.Message, logger base.SGLogger) *subChanges {
	return &subChanges{
		rq:     rq,
		logger: logger,
	}
}

func (s *subChanges) since(zeroValueCreator func() db.SequenceID) db.SequenceID {

	// Depending on the db sequence type, use correct zero sequence for since value
	s.sinceSequenceId = zeroValueCreator()

	if sinceStr, found := s.rq.Properties[BlipPropertySince]; found {
		var err error
		if s.sinceSequenceId, err = db.ParseSequenceIDFromJSON([]byte(sinceStr)); err != nil {
			s.logger.LogTo("Sync", "%s: Invalid sequence ID in 'since': %s", s.rq, sinceStr)
			s.sinceSequenceId = db.SequenceID{}
		}
	}

	return s.sinceSequenceId

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
		s.sinceSequenceId,
		s.continuous(),
		s.activeOnly(),
		s.filter(),
		channels,
	)
}
