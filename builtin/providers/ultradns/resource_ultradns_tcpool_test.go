package ultradns

import (
	"fmt"
	"testing"

	"github.com/Ensighten/udnssdk"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccUltradnsTcpool(t *testing.T) {
	var record udnssdk.RRSet
	domain := "ultradns.phinze.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUltradnsTcpoolDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckUltraDNSRecordTcpoolMinimal, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltraDNSRecordExists("ultradns_tcpool.tcpool-minimal", &record),
					testAccCheckUltraDNSRecordAttributes(&record),
					resource.TestCheckResourceAttr("ultradns_tcpool.tcpool-minimal", "name", "tcpool-minimal"),
					resource.TestCheckResourceAttr("ultradns_tcpool.tcpool-minimal", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_tcpool.tcpool-minimal", "rdata.0", "192.168.0.10"),
				),
			},
		},
	})
}

func testAccCheckUltradnsTcpoolDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*udnssdk.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ultradns_tcpool" {
			continue
		}

		k := udnssdk.RRSetKey{
			Zone: rs.Primary.Attributes["zone"],
			Name: rs.Primary.Attributes["name"],
			Type: rs.Primary.Attributes["type"],
		}

		_, err := client.RRSets.Select(k)

		if err == nil {
			return fmt.Errorf("Record still exists")
		}
	}

	return nil
}

const testAccCheckUltraDNSRecordTcpoolMinimal = `
resource "ultradns_tcpool" "tcpool-minimal" {
  zone = "%s"
  name = "tcpool-minimal"
  type = "A"

  rdata = ["192.168.0.10"]

  rdata_info {
    priority  = 1
    threshold = 1
  }
}
`

const testAccCheckUltraDNSRecordTcpoolMaximal = `
resource "ultradns_tcpool" "tcpool-maximal" {
  zone = "%s"
  name = "terraform-tcpool-maximal"
  type = "A"

  rdata = ["192.168.0.10"]

  description   = "traffic controller pool with all settings tuned"
  run_probes    = false
  act_on_probes = false
  order         = "ROUND_ROBIN"
  max_to_lb     = 2

  rdata_info {
    priority      = 1
    threshold     = 1
    state         = "ACTIVE"
    run_probes     = false
    failover_delay = 30
    weight        = 2
  }

  backup_record_rdata          = "192.168.0.11"
  backup_record_failover_delay = 30
}
`
