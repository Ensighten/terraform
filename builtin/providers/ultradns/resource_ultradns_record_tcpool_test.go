package ultradns

import (
	"fmt"
	"testing"

	"github.com/Ensighten/udnssdk"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUltraDNSRecord_TCPool(t *testing.T) {
	var record udnssdk.RRSet
	domain := "ultradns.phinze.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUltraDNSRecordDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckUltraDNSRecordTcpoolMinimal, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltraDNSRecordExists("ultradns_record.tcpool-minimal", &record),
					testAccCheckUltraDNSRecordAttributes(&record),
					resource.TestCheckResourceAttr("ultradns_record.tcpool-minimal", "name", "tcpool-minimal"),
					resource.TestCheckResourceAttr("ultradns_record.tcpool-minimal", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_record.tcpool-minimal", "rdata.0", "192.168.0.10"),
				),
			},
		},
	})
}

const testAccCheckUltraDNSRecordTcpoolMinimal = `
resource "ultradns_record" "tcpool-minimal" {
  zone = "%s"
  name = "tcpool-minimal"
  type = "A"

  rdata = ["192.168.0.10"]

  tcpool_profile {
    rdataInfo {
      priority  = 1
      threshold = 1
    }
  }
}
`

const testAccCheckUltraDNSRecordTcpoolMaximal = `
resource "ultradns_record" "tcpool-maximal" {
  zone = "%s"
  name = "terraform-tcpool-maximal"
  type = "A"

  rdata = ["192.168.0.10"]

  tcpool_profile {
    description = "traffic controller pool with all settings tuned"
    runProbes   = false
    actOnProbes = false
    order       = "ROUND_ROBIN"
    maxToLB     = 2

    rdataInfo {
      priority      = 1
      threshold     = 1
      state         = "ACTIVE"
      runProbes     = false
      failoverDelay = 30
      weight        = 2
    }

    backupRecord {
      rdata         = "192.168.0.11"
      failoverDelay = 30
    }
  }
}
`
