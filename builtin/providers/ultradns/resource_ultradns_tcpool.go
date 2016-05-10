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

func resourceUltradnsTcpool() *schema.Resource {
	return &schema.Resource{
		Create: resourceUltradnsTcpoolCreate,
		Read:   resourceUltradnsTcpoolRead,
		Update: resourceUltradnsTcpoolUpdate,
		Delete: resourceUltradnsTcpoolDelete,

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
			"rdata": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				// Valid: len(rdataInfo) == len(rdata)
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required
						"host": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						// Optional
						"failover_delay": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
							// Valid: 0-30
							// Units: Minutes
						},
						"priority": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						"run_probes": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"state": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "NORMAL",
						},
						"threshold": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						"weight": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  2,
							// Valid: i%2 == 0 && 2 <= i <= 100
						},
					},
				},
			},
			// Optional
			"ttl": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "3600",
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				// 0-255 char
			},
			"run_probes": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"act_on_probes": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"max_to_lb": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				// Valid: 0 <= i <= len(rdata)
			},
			"backup_record_rdata": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				// Valid: IPv4 address or CNAME
			},
			"backup_record_failover_delay": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				// Valid: 0-30
				// Units: Minutes
			},
			// Computed
			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// CRUD Operations

func resourceUltradnsTcpoolCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newRRSetResourceFromTcpool(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] ultradns_tcpool create: %+v", r)
	_, err = client.RRSets.Create(r.RRSetKey(), r.RRSet())
	if err != nil {
		return fmt.Errorf("ultradns_tcpool create failed: %v", err)
	}

	d.SetId(r.ID())
	log.Printf("[INFO] ultradns_tcpool.id: %v", d.Id())

	return resourceUltradnsTcpoolRead(d, meta)
}

func resourceUltradnsTcpoolRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newRRSetResourceFromTcpool(d)
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
				return fmt.Errorf("ultradns_tcpool not found: %v", err)
			}
		}
		return fmt.Errorf("ultradns_tcpool not found: %v", err)
	}
	rec := rrsets[0]
	return populateTcpoolFromRRSet(rec, d)
}

func resourceUltradnsTcpoolUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newRRSetResourceFromTcpool(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] ultradns_tcpool update: %+v", r)
	_, err = client.RRSets.Update(r.RRSetKey(), r.RRSet())
	if err != nil {
		return fmt.Errorf("ultradns_tcpool update failed: %v", err)
	}

	return resourceUltradnsTcpoolRead(d, meta)
}

// Resource Helpers

func resourceUltradnsTcpoolDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newRRSetResourceFromTcpool(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] ultradns_tcpool delete: %+v", r)
	_, err = client.RRSets.Delete(r.RRSetKey())
	if err != nil {
		return fmt.Errorf("ultradns_tcpool delete failed: %v", err)
	}

	return nil
}

func newRRSetResourceFromTcpool(d *schema.ResourceData) (rRSetResource, error) {
	r := rRSetResource{
		Zone:      d.Get("zone").(string),
		OwnerName: d.Get("name").(string),
		// "The only valid rrtype value for SiteBacker or Traffic Controller pools is A"
		// per https://portal.ultradns.com/static/docs/REST-API_User_Guide.pdf
		RRType: "A",
	}

	rDataRaw := d.Get("rdata").(*schema.Set).List()
	r.RData = expandRdataHosts(rDataRaw)

	profile := udnssdk.TCPoolProfile{
		Context:   udnssdk.TCPoolSchema,
		RDataInfo: expandRdataInfos(rDataRaw),
	}

	// Optional attributes
	if attr, ok := d.GetOk("ttl"); ok {
		r.TTL, _ = strconv.Atoi(attr.(string))
	}
	if attr, ok := d.GetOk("description"); ok {
		profile.Description = attr.(string)
	}
	if attr, ok := d.GetOk("run_probes"); ok {
		profile.RunProbes = attr.(bool)
	}
	if attr, ok := d.GetOk("act_on_probes"); ok {
		profile.ActOnProbes = attr.(bool)
	}
	if attr, ok := d.GetOk("max_to_lb"); ok {
		profile.MaxToLB = attr.(int)
	}
	if attr, ok := d.GetOk("backup_record_rdata"); ok {
		profile.BackupRecord = udnssdk.BackupRecord{
			RData: attr.(string),
		}
		if attr, ok := d.GetOk("backup_record_failover_delay"); ok {
			profile.BackupRecord.FailoverDelay = attr.(int)
		}
	}

	s, err := json.Marshal(profile)
	if err != nil {
		return r, fmt.Errorf("ultradns_tcpool profile marshal error: %+v", err)
	}
	r.Profile = &udnssdk.StringProfile{Profile: string(s)}

	return r, nil
}

func expandRdataHosts(configured []interface{}) []string {
	hs := make([]string, len(configured))
	for _, rRaw := range configured {
		data := rRaw.(map[string]interface{})
		h := data["host"].(string)
		hs = append(hs, h)
	}
	return hs
}

func expandRdataInfos(configured []interface{}) []udnssdk.SBRDataInfo {
	rdataInfos := make([]udnssdk.SBRDataInfo, len(configured))
	for _, rRaw := range configured {
		data := rRaw.(map[string]interface{})
		r := udnssdk.SBRDataInfo{
			FailoverDelay: data["failover_delay"].(int),
			Priority:      data["priority"].(int),
			RunProbes:     data["run_probes"].(bool),
			State:         data["state"].(string),
			Threshold:     data["threshold"].(int),
			Weight:        data["weight"].(int),
		}
		rdataInfos = append(rdataInfos, r)
	}
	return rdataInfos
}

func populateTcpoolFromRRSet(r udnssdk.RRSet, d *schema.ResourceData) error {
	zone := d.Get("zone")
	// ttl
	d.Set("ttl", r.TTL)
	// rdata
	err := d.Set("rdata", r.RData)
	if err != nil {
		return fmt.Errorf("ultradns_tcpool.rdata set failed: %#v", err)
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
		c := r.Profile.Context()
		if c != udnssdk.TCPoolSchema {
			return fmt.Errorf("ultradns_tcpool profile has unknown type %s\n", c)
		}
		pRaw := r.Profile.GetProfileObject()
		p, ok := pRaw.(udnssdk.TCPoolProfile)
		if ok {
			return fmt.Errorf("ultradns_tcpool profile could not be unmarshalled\n")
		}
		err := d.Set("description", p.Description)
		if err != nil {
			return fmt.Errorf("ultradns_tcpool.description set failed: %#v", err)
		}
		err = d.Set("run_probes", p.RunProbes)
		if err != nil {
			return fmt.Errorf("ultradns_tcpool.run_probes set failed: %#v", err)
		}
		err = d.Set("act_on_probes", p.ActOnProbes)
		if err != nil {
			return fmt.Errorf("ultradns_tcpool.act_on_probes set failed: %#v", err)
		}
		err = d.Set("max_to_lb", p.MaxToLB)
		if err != nil {
			return fmt.Errorf("ultradns_tcpool.max_to_lb set failed: %#v", err)
		}
		err = d.Set("backup_record_rdata", p.BackupRecord.RData)
		if err != nil {
			return fmt.Errorf("ultradns_tcpool.backup_record_rdata set failed: %#v", err)
		}
		err = d.Set("backup_record_failover_delay", p.BackupRecord.FailoverDelay)
		if err != nil {
			return fmt.Errorf("ultradns_tcpool.backup_record_failover_delay set failed: %#v", err)
		}
	}
	return nil
}
