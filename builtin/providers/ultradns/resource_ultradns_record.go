package ultradns

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Ensighten/udnssdk"
	"github.com/hashicorp/terraform/helper/schema"
)

func newRRSetResource(d *schema.ResourceData) (rRSetResource, error) {
	r := rRSetResource{}

	// TODO: return error if required attributes aren't ok

	if attr, ok := d.GetOk("name"); ok {
		r.OwnerName = attr.(string)
	}

	if attr, ok := d.GetOk("type"); ok {
		r.RRType = attr.(string)
	}

	if attr, ok := d.GetOk("zone"); ok {
		r.Zone = attr.(string)
	}

	if attr, ok := d.GetOk("rdata"); ok {
		rdata := attr.([]interface{})
		r.RData = make([]string, len(rdata))
		for i, j := range rdata {
			r.RData[i] = j.(string)
		}
	}

	if attr, ok := d.GetOk("ttl"); ok {
		r.TTL, _ = strconv.Atoi(attr.(string))
	}

	if attr, ok := d.GetOk("string_profile"); ok {
		r.Profile = &udnssdk.StringProfile{Profile: attr.(string)}
	}

	for k, schema := range profileAttrSchemaMap {
		if attr, ok := d.GetOk(k); ok {
			poolProfile := attr.(map[string]interface{})
			if len(poolProfile) != 0 {
				poolProfile["@context"] = schema
				s, err := json.Marshal(poolProfile)
				if err != nil {
					return r, fmt.Errorf("ultradns_record string_profile marshal error: %+v", err)
				}
				r.Profile = &udnssdk.StringProfile{Profile: string(s)}
				break
			}
		}
	}

	return r, nil
}

func populateResourceDataFromRRSet(r udnssdk.RRSet, d *schema.ResourceData) error {
	zone := d.Get("zone")
	// ttl
	d.Set("ttl", r.TTL)
	// rdata
	err := d.Set("rdata", r.RData)
	if err != nil {
		return fmt.Errorf("ultradns_record.rdata set failed: %#v", err)
	}
	// hostname
	if r.OwnerName == "" {
		d.Set("hostname", zone)
	} else {
		if strings.HasSuffix(r.OwnerName, ".") {
			d.Set("hostname", r.OwnerName)
		} else {
			d.Set("hostname", fmt.Sprintf("%s.%s", r.OwnerName, zone))
		}
	}
	// *_profile
	if r.Profile != nil {
		d.Set("string_profile", r.Profile.Profile)
		// TODO: use udnssdk.StringProfile.GetProfileObject()
		var p map[string]interface{}
		err = json.Unmarshal([]byte(r.Profile.Profile), &p)
		if err != nil {
			return err
		}
		c := r.Profile.Context()
		switch c {
		case udnssdk.DirPoolSchema:
			d.Set("dirpool_profile", p)
		case udnssdk.RDPoolSchema:
			d.Set("rdpool_profile", p)
		case udnssdk.SBPoolSchema:
			d.Set("sbpool_profile", p)
		case udnssdk.TCPoolSchema:
			d.Set("tcpool_profile", p)
		default:
			return fmt.Errorf("ultradns_record profile has unknown type %s\n", c)
		}
	}
	return nil
}

func schemaDirPoolProfile() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		ConflictsWith: []string{
			"string_profile",
			"rdpool_profile",
			"sbpool_profile",
			"tcpool_profile",
		},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"description": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  "RD Pool Profile created by Terraform",
				},
				"conflictResolve": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  "GEO",
				},
				"rdataInfo": &schema.Schema{
					Type:     schema.TypeSet,
					Optional: true,
					Elem:     schemaDirPoolRDataInfo(),
				},
				"noResponse": schemaDirPoolRDataInfo(),
			},
		},
	}
}

func schemaDirPoolRDataInfo() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"allNonConfigured": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  false,
				},
				"geoInfo": &schema.Schema{
					Type:     schema.TypeSet,
					Optional: true,
					Elem:     schemaDirPoolGeoInfo(),
				},
				"ipInfo": &schema.Schema{
					Type:     schema.TypeSet,
					Optional: true,
					Elem:     schemaDirPoolIPInfo(),
				},
			},
		},
	}
}

func schemaDirPoolGeoInfo() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": &schema.Schema{
					Type:     schema.TypeString,
					Optional: false,
				},
				"isAccountLevel": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  false,
				},
				"codes": &schema.Schema{
					Type:     schema.TypeSet,
					Optional: true,
					Elem:     schema.TypeString,
				},
			},
		},
	}
}

func schemaDirPoolIPInfo() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": &schema.Schema{
					Type:     schema.TypeString,
					Optional: false,
				},
				"isAccountLevel": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  false,
				},
				"ips": &schema.Schema{
					Type:     schema.TypeSet,
					Optional: true,
					Elem:     schemaIPAddrDTO(),
				},
			},
		},
	}
}

func schemaIPAddrDTO() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"start": &schema.Schema{
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"cidr", "address"},
				},
				"end": &schema.Schema{
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"cidr", "address"},
				},
				"cidr": &schema.Schema{
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"start", "end", "address"},
				},
				"address": &schema.Schema{
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"start", "end", "cidr"},
				},
			},
		},
	}
}

func schemaSBPoolProfile() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		ConflictsWith: []string{
			"string_profile",
			"dirpool_profile",
			"rdpool_profile",
			"tcpool_profile",
		},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"description": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  "RD Pool Profile created by Terraform",
				},
				"conflictResolve": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  "GEO",
				},
				"rdataInfo": &schema.Schema{
					Type:     schema.TypeSet,
					Optional: true,
					Elem:     schemaDirPoolRDataInfo(),
				},
				"noResponse": schemaDirPoolRDataInfo(),
			},
		},
	}
}

func schemaTCPoolProfile() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,

		ConflictsWith: []string{
			"string_profile",
			"dirpool_profile",
			"sbpool_profile",
			"rdpool_profile",
		},
	}
}

func schemaRDPoolProfile() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		ConflictsWith: []string{
			"string_profile",
			"dirpool_profile",
			"sbpool_profile",
			"tcpool_profile",
		},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"@context": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  udnssdk.RDPoolSchema,
				},
				"order": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  "ROUND_ROBIN",
				},
				"description": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  "RD Pool Profile created by Terraform",
				},
			},
		},
	}
}

func resourceUltraDNSRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceUltraDNSRecordCreate,
		Read:   resourceUltraDNSRecordRead,
		Update: resourceUltraDNSRecordUpdate,
		Delete: resourceUltraDNSRecordDelete,

		Schema: map[string]*schema.Schema{
			// Required
			"zone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rdata": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// Optional
			"ttl": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "3600",
			},
			"string_profile": &schema.Schema{
				Type: schema.TypeString,
				ConflictsWith: []string{
					"dirpool_profile",
					"sbpool_profile",
					"tcpool_profile",
					"rdpool_profile",
				},
				Optional: true,
			},
			"rdpool_profile":  schemaRDPoolProfile(),
			"dirpool_profile": schemaDirPoolProfile(),
			"sbpool_profile":  schemaSBPoolProfile(),
			"tcpool_profile":  schemaTCPoolProfile(),
			// Computed
			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// CRUD Operations

func resourceUltraDNSRecordCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newRRSetResource(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] ultradns_record create: %+v", r)
	_, err = client.RRSets.Create(r.RRSetKey(), r.RRSet())
	if err != nil {
		return fmt.Errorf("ultradns_record create failed: %v", err)
	}

	d.SetId(r.ID())
	log.Printf("[INFO] ultradns_record.id: %v", d.Id())

	return resourceUltraDNSRecordRead(d, meta)
}

func resourceUltraDNSRecordRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newRRSetResource(d)
	if err != nil {
		return err
	}

	rrsets, err := client.RRSets.Select(r.RRSetKey())
	if err != nil {
		uderr, ok := err.(*udnssdk.ErrorResponseList)
		if ok {
			for _, r := range uderr.Responses {
				// 70002 means Records Not Found
				if r.ErrorCode == 70002 {
					d.SetId("")
					return nil
				}
				return fmt.Errorf("ultradns_record not found: %v", err)
			}
		}
		return fmt.Errorf("ultradns_record not found: %v", err)
	}
	rec := rrsets[0]
	return populateResourceDataFromRRSet(rec, d)
}

func resourceUltraDNSRecordUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newRRSetResource(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] ultradns_record update: %+v", r)
	_, err = client.RRSets.Update(r.RRSetKey(), r.RRSet())
	if err != nil {
		return fmt.Errorf("ultradns_record update failed: %v", err)
	}

	return resourceUltraDNSRecordRead(d, meta)
}

func resourceUltraDNSRecordDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newRRSetResource(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] ultradns_record delete: %+v", r)
	_, err = client.RRSets.Delete(r.RRSetKey())
	if err != nil {
		return fmt.Errorf("ultradns_record delete failed: %v", err)
	}

	return nil
}

// Conversion helper functions
