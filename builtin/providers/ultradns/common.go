package ultradns

import (
	"encoding/json"
	"fmt"

	"github.com/Ensighten/udnssdk"
)

// Conversion helper functions
type rRSetResource struct {
	OwnerName string
	RRType    string
	RData     []string
	TTL       int
	Profile   *udnssdk.StringProfile
	Zone      string
}

// profileAttrSchemaMap is a map from each ultradns_tcpool attribute name onto its respective ProfileSchema URI
var profileAttrSchemaMap = map[string]udnssdk.ProfileSchema{
	"dirpool_profile": udnssdk.DirPoolSchema,
	"rdpool_profile":  udnssdk.RDPoolSchema,
	"sbpool_profile":  udnssdk.SBPoolSchema,
	"tcpool_profile":  udnssdk.TCPoolSchema,
}

func (r rRSetResource) RRSetKey() udnssdk.RRSetKey {
	return udnssdk.RRSetKey{
		Zone: r.Zone,
		Type: r.RRType,
		Name: r.OwnerName,
	}
}

func (r rRSetResource) RRSet() udnssdk.RRSet {
	return udnssdk.RRSet{
		OwnerName: r.OwnerName,
		RRType:    r.RRType,
		RData:     r.RData,
		TTL:       r.TTL,
	}
}

func (r rRSetResource) ID() string {
	return fmt.Sprintf("%s.%s", r.OwnerName, r.Zone)
}

// TODO: move this to udnssdk (r *RRSetSet) SetProfile(p interface{}) error
func (r rRSetResource) SetProfile(p interface{}) error {
	s, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("ultradns rRSetResource.SetProfile(): profile marshal error: %+v", err)
	}
	r.Profile = &udnssdk.StringProfile{Profile: string(s)}
	return nil
}
