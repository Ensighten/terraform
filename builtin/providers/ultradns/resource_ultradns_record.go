package ultradns

import (
	"encoding/json"
	"fmt"
	"github.com/Ensighten/udnssdk"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strconv"
	"strings"
)

func schemaSBPoolProfile() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		ConflictsWith: []string{
			"dirpool_profile",
			"rdpool_profile",
			"tcpool_profile",
			"string_profile",
			"map_profile",
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

func schemaDirPoolProfile() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeMap,
		Optional:      true,
		ConflictsWith: []string{"rdpool_profile", "sbpool_profile", "tcpool_profile", "string_profile", "map_profile"},
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

		ConflictsWith: []string{"dirpool_profile", "sbpool_profile", "rdpool_profile", "string_profile", "map_profile"},
	}
}
func schemaRDPoolProfile() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeMap,
		Optional:      true,
		ConflictsWith: []string{"dirpool_profile", "sbpool_profile", "tcpool_profile", "string_profile", "map_profile"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"@context": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  "http://schemas.ultradns.com/RDPool.jsonschema",
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

			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
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
			"ttl": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "3600",
			},
			"rdpool_profile":  schemaRDPoolProfile(),
			"dirpool_profile": schemaDirPoolProfile(),
			"sbpool_profile":  schemaSBPoolProfile(),
			"tcpool_profile":  schemaTCPoolProfile(),
			"map_profile": &schema.Schema{
				Type:          schema.TypeMap,
				ConflictsWith: []string{"dirpool_profile", "sbpool_profile", "tcpool_profile", "rdpool_profile", "string_profile"},
				Optional:      true,
			},
			"string_profile": &schema.Schema{
				Type:          schema.TypeString,
				ConflictsWith: []string{"dirpool_profile", "sbpool_profile", "tcpool_profile", "rdpool_profile", "map_profile"},
				Optional:      true,
			},
		},
	}
}

type RRSetResource struct {
	OwnerName string
	RRType    string
	RData     []string
	TTL       int
	Profile   *udnssdk.StringProfile
	Zone      string
}

func NewRRSetResource(d *schema.ResourceData) (RRSetResource, error) {
	r := RRSetResource{}

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

	newProfileStr := d.Get("string_profile").(string)
	if newProfileStr != "" {
		r.Profile = &udnssdk.StringProfile{
			Profile: newProfileStr,
		}
	}

	profilelist := map[string]string{
		"rdpool_profile":  "http://schemas.ultradns.com/RDPool.jsonschema",
		"sbpool_profile":  "http://schemas.ultradns.com/SBPool.jsonschema",
		"tcpool_profile":  "http://schemas.ultradns.com/TCPool.jsonschema",
		"dirpool_profile": "http://schemas.ultradns.com/DirPool.jsonschema",
	}
	for key, schemaURL := range profilelist {
		firstValidation := d.Get(key)
		if firstValidation == nil {
			continue
		}
		poolProfile := firstValidation.(map[string]interface{})
		if len(poolProfile) != 0 {
			poolProfile["@context"] = schemaURL
			x, err := json.Marshal(poolProfile)
			if err != nil {
				return r, fmt.Errorf("[ERROR] poolProfile Marshal error: %+v", err)
			}
			r.Profile = &udnssdk.StringProfile{
				Profile: string(x),
			}
			break
		}
	}

	return r, nil
}

func (r RRSetResource) RRSetKey() udnssdk.RRSetKey {
	return udnssdk.RRSetKey{
		Zone: r.Zone,
		Type: r.RRType,
		Name: r.OwnerName,
	}
}

func (r RRSetResource) RRSet() udnssdk.RRSet {
	return udnssdk.RRSet{
		OwnerName: r.OwnerName,
		RRType:    r.RRType,
		RData:     r.RData,
		TTL:       r.TTL,
	}
}

func (r RRSetResource) ID() string {
	return fmt.Sprintf("%s.%s", r.OwnerName, r.Zone)
}

func resourceUltraDNSRecordCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := NewRRSetResource(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] UltraDNS RRSet create configuration: %#v", r.RRSet())
	_, err = client.RRSets.Create(r.RRSetKey(), r.RRSet())
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to create UltraDNS RRSet: %s", err)
	}

	d.SetId(r.ID())
	log.Printf("[INFO] record ID: %s", d.Id())

	return resourceUltraDNSRecordRead(d, meta)
}

func resourceUltraDNSRecordRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := NewRRSetResource(d)
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
				} else {
					return fmt.Errorf("[ERROR] Couldn't find UltraDNS RRSet: %s", err)
				}
			}
		} else {
			return fmt.Errorf("[ERROR] Couldn't find UltraDNS RRSet: %s", err)
		}
	}
	rec := rrsets[0]
	// ttl
	d.Set("ttl", rec.TTL)
	// rdata
	err = d.Set("rdata", rec.RData)
	if err != nil {
		return fmt.Errorf("[DEBUG] Error setting records: %#v", err)
	}
	// hostname
	if rec.OwnerName == "" {
		d.Set("hostname", r.Zone)
	} else {
		if strings.HasSuffix(rec.OwnerName, ".") {
			d.Set("hostname", rec.OwnerName)
		} else {
			d.Set("hostname", fmt.Sprintf("%s.%s", rec.OwnerName, r.Zone))
		}
	}
	// *_profile
	if rec.Profile != nil {
		t := rec.Profile.GetType()
		d.Set("string_profile", rec.Profile.Profile)
		var p map[string]interface{}
		err = json.Unmarshal([]byte(rec.Profile.Profile), &p)
		if err != nil {
			return err
		}
		typ := strings.Split(t, "/")
		switch typ[len(typ)-1] {
		case "DirPool.jsonschema":
			d.Set("dirpool_profile", p)
		case "RDPool.jsonschema":
			d.Set("rdpool_profile", p)
		case "TCPool.jsonschema":
			d.Set("tcpool_profile", p)
		case "SBPool.jsonschema":
			d.Set("sbpool_profile", p)
		default:
			return fmt.Errorf("[DEBUG] Unknown Type %s\n", t)
		}
	}
	return nil
}

func resourceUltraDNSRecordUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := NewRRSetResource(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] UltraDNS RRSet update configuration: %#v", r.RRSet())
	_, err = client.RRSets.Update(r.RRSetKey(), r.RRSet())
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to update UltraDNS RRSet: %s", err)
	}

	return resourceUltraDNSRecordRead(d, meta)
}

func resourceUltraDNSRecordDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := NewRRSetResource(d)
	if err != nil {
		return err
	}

	_, err = client.RRSets.Delete(r.RRSetKey())

	if err != nil {
		return fmt.Errorf("[ERROR] Error deleting UltraDNS RRSet: %s", err)
	}

	return nil
}
