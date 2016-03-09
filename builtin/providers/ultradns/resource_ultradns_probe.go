package ultradns

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Ensighten/udnssdk"
	"github.com/hashicorp/terraform/helper/schema"
)

var typeToAttrKeyMap = map[string]string{
	"HTTP":      "http_probe",
	"PING":      "ping_probe",
	"FTP":       "ftp_probe",
	"SMTP":      "smtp_probe",
	"SMTP_SEND": "smtpsend_probe",
	"DNS":       "dns_probe",
}

type probeResource struct {
	Name string
	Zone string
	ID   string

	Agents     []string
	Interval   string
	PoolRecord string
	Threshold  int
	Type       string
	Details    *udnssdk.ProbeDetailsDTO
}

func newProbeResource(d *schema.ResourceData) (probeResource, error) {
	p := probeResource{}
	// zoneName
	p.Zone = d.Get("zoneName").(string)
	// ownerName
	p.Name = d.Get("ownerName").(string)
	// id
	p.ID = d.Id()

	p.PoolRecord = d.Get("poolRecord").(string)
	p.Type = d.Get("type").(string)
	p.Interval = d.Get("interval").(string)
	p.Threshold = d.Get("threshold").(int)

	// agents
	oldagents, ok := d.GetOk("agents")
	if !ok {
		return p, fmt.Errorf("ultradns_probe.agents not ok")
	}
	for _, e := range oldagents.([]interface{}) {
		p.Agents = append(p.Agents, e.(string))
	}

	// details
	// TODO: validate p.Type is in typeToAttrKeyMap.Keys
	typeAttrKey := typeToAttrKeyMap[p.Type]
	attr, ok := d.GetOk(typeAttrKey)
	if !ok {
		return p, fmt.Errorf("ultradns_probe.%s not ok", typeAttrKey)
	}
	probeset := attr.(*schema.Set)
	var probedetails map[string]interface{}
	probedetails = probeset.List()[0].(map[string]interface{})
	// Convert limits from flattened set format to mapping.
	limits := map[string]interface{}{}
	for _, limit := range probedetails["limits"].([]interface{}) {
		l := limit.(map[string]interface{})
		name := l["name"].(string)
		limits[name] = map[string]interface{}{
			"warning":  l["warning"],
			"critical": l["critical"],
			"fail":     l["fail"],
		}
	}
	probedetails["limits"] = limits

	p.Details = &udnssdk.ProbeDetailsDTO{
		Detail: probedetails,
	}
	return p, nil
}

func (r probeResource) Key() udnssdk.ProbeKey {
	k := udnssdk.ProbeKey{
		Name: r.Name,
		Zone: r.Zone,
		ID:   r.ID,
	}
	return k
}

func (r probeResource) ProbeInfoDTO() udnssdk.ProbeInfoDTO {
	p := udnssdk.ProbeInfoDTO{
		ID:         r.ID,
		PoolRecord: r.PoolRecord,
		ProbeType:  r.Type,
		Interval:   r.Interval,
		Agents:     r.Agents,
		Threshold:  r.Threshold,
		Details:    r.Details,
	}
	return p
}

func populateResourceDataFromProbe(p udnssdk.ProbeInfoDTO, d *schema.ResourceData) error {
	err := p.Details.Populate(p.ProbeType)
	if err != nil {
		return fmt.Errorf("Could not populate probe details: %#v", err)
	}
	// poolRecord
	err = d.Set("poolRecord", p.PoolRecord)
	if err != nil {
		return fmt.Errorf("Error setting poolRecord: %#v", err)
	}
	// interval
	err = d.Set("interval", p.Interval)
	if err != nil {
		return fmt.Errorf("Error setting interval: %#v", err)
	}
	// type
	err = d.Set("type", p.ProbeType)
	if err != nil {
		return fmt.Errorf("Error setting type: %#v", err)
	}
	// agents
	err = d.Set("agents", p.Agents)
	if err != nil {
		return fmt.Errorf("Error setting agents: %#v", err)
	}
	// threshold
	err = d.Set("threshold", p.Threshold)
	if err != nil {
		return fmt.Errorf("Error setting threshold: %#v", err)
	}
	// id
	d.SetId(p.ID)
	// details
	if p.Details != nil {
		var dp map[string]interface{}
		err = json.Unmarshal(p.Details.GetData(), &dp)
		if err != nil {
			return err
		}

		err = d.Set(typeToAttrKeyMap[p.ProbeType], dp)
		if err != nil {
			return fmt.Errorf("Error setting details: %#v", err)
		}

	}
	return nil
}

/*
func schemaTransaction() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: false,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"method": &schema.Schema{
					Type:     schema.TypeString,
					Optional: false,
				},
				"url": &schema.Schema{
					Type:     schema.TypeString,
					Optional: false,
				},
				"transmittedData": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"followRedirects": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"limits": &schema.Schema{
					Type:     schema.TypeSet,
					Optional: false,
					Elem:     resourceLimits(),
				},
			},
		},
	}
}

/*
func schemaPingProbeLimits() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"lossPercent": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     resourceLimits(),
				},
				"total": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     resourceLimits(),
				},
				"average": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     resourceLimits(),
				},
				"run": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     resourceLimits(),
				},
				"avgRun": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     resourceLimits(),
				},
			},
		},
	}
}
func resourcePingProbeLimits() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"lossPercent": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     resourceLimits(),
			},
			"total": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     resourceLimits(),
			},
			"average": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     resourceLimits(),
			},
			"run": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     resourceLimits(),
			},
			"avgRun": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     resourceLimits(),
			},
		},
	}
}
*/
func resourceLimits() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"warning": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"critical": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"fail": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

/*
func schemaHTTPProbe() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"transactions": schemaTransaction(),
				//"totalLimits":  schemaLimits(),
			},
		},
	}
}
func schemaSMTPProbe() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"port": &schema.Schema{
					Type:     schema.TypeInt,
					Optional: true,
				},
				"limits": &schema.Schema{
					Type:     schema.TypeMap,
					Required: true,
					Elem:     resourceLimits(),
				},
			},
		},
	}
}
func schemaSMTPSENDProbe() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"port": &schema.Schema{
					Type:     schema.TypeInt,
					Optional: true,
				},
				"from": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
				"to": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},

				"message": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
				"limits": &schema.Schema{
					Type:     schema.TypeMap,
					Required: true,
					Elem:     resourceLimits(),
				},
			},
		},
	}
}

func schemaTCPProbe() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"port": &schema.Schema{
					Type:     schema.TypeInt,
					Optional: true,
				},
				"controlIP": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"limits": &schema.Schema{
					Type:     schema.TypeMap,
					Required: true,
					Elem:     resourceLimits(),
				},
			},
		},
	}
}

func schemaFTPProbe() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"port": &schema.Schema{
					Type:     schema.TypeInt,
					Optional: true,
				},
				"passiveMode": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"username": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"password": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"path": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
				"limits": &schema.Schema{
					Type:     schema.TypeMap,
					Required: true,
					Elem:     resourceLimits(),
				},
			},
		},
	}
}
*/
func schemaPingProbe() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"packets": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"packetSize": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"limits": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     resourceLimits(),
			},
		},
	}

}

/*
func schemaDNSProbe() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"port": &schema.Schema{
					Type:     schema.TypeInt,
					Optional: true,
				},
				"tcpOnly": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"type": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"ownerName": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"limits": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     resourceLimits(),
				},
			},
		},
	}
}
*/
func resourceUltraDNSProbe() *schema.Resource {
	return &schema.Resource{
		Create: resourceUltraDNSProbeCreate,
		Read:   resourceUltraDNSProbeRead,
		Update: resourceUltraDNSProbeUpdate,
		Delete: resourceUltraDNSProbeDelete,

		Schema: map[string]*schema.Schema{
			// Required
			"ownerName": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ownerType": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"zoneName": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"poolRecord": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"interval": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"agents": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"threshold": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			// Optional
			"ping_probe": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				//ConflictsWith: []string{"http_probe", "smtp_probe", "dns_probe", "smtpsend_probe"},
				Elem: schemaPingProbe(),
			},
			// "http_probe": &schema.Schema{
			// 	Type:          schema.TypeSet,
			// 	Optional:      true,
			// 	ConflictsWith: []string{"dns_probe", "ping_probe", "smtp_probe", "smtpsend_probe"},
			// 	Elem:          schemaHTTPProbe(),
			// },
			// "dns_probe": &schema.Schema{
			// 	Type:          schema.TypeSet,
			// 	Optional:      true,
			// 	ConflictsWith: []string{"http_probe", "ping_probe", "smtp_probe", "smtpsend_probe"},
			// 	Elem:          schemaDNSProbe(),
			// },
			// "smtpsend_probe": &schema.Schema{
			// 	Type:          schema.TypeSet,
			// 	Optional:      true,
			// 	ConflictsWith: []string{"http_probe", "ping_probe", "smtp_probe", "dns_probe"},
			// 	Elem:          schemaDNSProbe(),
			// },
			// "smtp_probe": &schema.Schema{
			// 	Type:          schema.TypeSet,
			// 	Optional:      true,
			// 	ConflictsWith: []string{"http_probe", "ping_probe", "dns_probe", "smtpsend_probe"},
			// 	Elem:          schemaDNSProbe(),
			// },
			// Computed
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUltraDNSProbeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newProbeResource(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] ultradns_probe create: %#v", r)
	resp, err := client.Probes.Create(r.Key().RRSetKey(), r.ProbeInfoDTO())
	if err != nil {
		return fmt.Errorf("ultradns_probe create failed: %s", err)
	}

	uri := resp.Header.Get("Location")
	d.Set("uri", uri)
	id := resp.Header.Get("ID")
	d.SetId(id)
	log.Printf("[INFO] ultradns_probe.id: %s", d.Id())

	return resourceUltraDNSProbeRead(d, meta)
}

func resourceUltraDNSProbeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)
	r, err := newProbeResource(d)
	if err != nil {
		return err
	}

	probe, _, err := client.Probes.Find(r.Key())
	if err != nil {
		uderr, ok := err.(*udnssdk.ErrorResponseList)
		if ok {
			for _, r := range uderr.Responses {
				// 70002 means Probes Not Found
				if r.ErrorCode == 70002 {
					d.SetId("")
					return nil
				}
				return fmt.Errorf("ultradns_probe not found: %s", err)
			}
		}
		return fmt.Errorf("ultradns_probe not found: %s", err)
	}
	return populateResourceDataFromProbe(probe, d)
}

func resourceUltraDNSProbeUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newProbeResource(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] ultradns_probe update: %#v", r)
	_, err = client.Probes.Update(r.Key(), r.ProbeInfoDTO())
	if err != nil {
		return fmt.Errorf("ultradns_probe update failed: %s", err)
	}

	return resourceUltraDNSProbeRead(d, meta)
}

func resourceUltraDNSProbeDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newProbeResource(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] ultradns_probe delete: %#v", r)
	_, err = client.Probes.Delete(r.Key())
	if err != nil {
		return fmt.Errorf("ultradns_probe delete failed: %s", err)
	}

	return nil
}
