package cloudstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

func TestAccCloudStackInstance_basic(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCloudStackInstance_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					testAccCheckCloudStackInstanceAttributes(&instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "user_data", "0cf3dcdc356ec8369494cb3991985ecd5296cdd5"),
					// testAccCheckResourceTags(&instance),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_update(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCloudStackInstance_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					testAccCheckCloudStackInstanceAttributes(&instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "user_data", "0cf3dcdc356ec8369494cb3991985ecd5296cdd5"),
				),
			},

			resource.TestStep{
				Config: testAccCloudStackInstance_renameAndResize,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					testAccCheckCloudStackInstanceRenamedAndResized(&instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "name", "terraform-updated"),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "display_name", "terraform-updated"),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "service_offering", "Medium Instance"),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_fixedIP(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCloudStackInstance_fixedIP,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "ip_address", "10.1.1.123"),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_keyPair(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCloudStackInstance_keyPair,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "keypair", "terraform-test-keypair"),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_project(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCloudStackInstance_project,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "project", "terraform"),
				),
			},
		},
	})
}

func testAccCheckCloudStackInstanceExists(
	n string, instance *cloudstack.VirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)
		vm, _, err := cs.VirtualMachine.GetVirtualMachineByID(rs.Primary.ID)

		if err != nil {
			return err
		}

		if vm.Id != rs.Primary.ID {
			return fmt.Errorf("Instance not found")
		}

		*instance = *vm

		return nil
	}
}

func testAccCheckCloudStackInstanceAttributes(
	instance *cloudstack.VirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if instance.Name != "terraform-test" {
			return fmt.Errorf("Bad name: %s", instance.Name)
		}

		if instance.Displayname != "terraform-test" {
			return fmt.Errorf("Bad display name: %s", instance.Displayname)
		}

		if instance.Serviceofferingname != "Small Instance" {
			return fmt.Errorf("Bad service offering: %s", instance.Serviceofferingname)
		}

		if instance.Templatename != "CentOS 5.6 (64-bit) no GUI (Simulator)" {
			return fmt.Errorf("Bad template: %s", instance.Templatename)
		}

		return nil
	}
}

func testAccCheckCloudStackInstanceRenamedAndResized(
	instance *cloudstack.VirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if instance.Name != "terraform-updated" {
			return fmt.Errorf("Bad name: %s", instance.Name)
		}

		if instance.Displayname != "terraform-updated" {
			return fmt.Errorf("Bad display name: %s", instance.Displayname)
		}

		if instance.Serviceofferingname != "Medium Instance" {
			return fmt.Errorf("Bad service offering: %s", instance.Serviceofferingname)
		}

		return nil
	}
}

func testAccCheckCloudStackInstanceDestroy(s *terraform.State) error {
	cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudstack_instance" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, _, err := cs.VirtualMachine.GetVirtualMachineByID(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Virtual Machine %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccCloudStackInstance_basic = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  zone = "Sandbox-simulator"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-test"
  display_name = "terraform-test"
  service_offering= "Small Instance"
  network_id = "${cloudstack_network.foo.id}"
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = "Sandbox-simulator"
  user_data = "foobar\nfoo\nbar"
  expunge = true
	#tags = {
	#  terraform-tag = "true"
	#}
}`

const testAccCloudStackInstance_renameAndResize = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  zone = "Sandbox-simulator"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-updated"
  display_name = "terraform-updated"
  service_offering= "Medium Instance"
  network_id = "${cloudstack_network.foo.id}"
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = "Sandbox-simulator"
  user_data = "foobar\nfoo\nbar"
  expunge = true
}`

const testAccCloudStackInstance_fixedIP = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  zone = "Sandbox-simulator"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-test"
  display_name = "terraform-test"
  service_offering= "Small Instance"
  network_id = "${cloudstack_network.foo.id}"
  ip_address = "10.1.1.123"
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = "Sandbox-simulator"
  expunge = true
}`

const testAccCloudStackInstance_keyPair = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  zone = "Sandbox-simulator"
}

resource "cloudstack_ssh_keypair" "foo" {
  name = "terraform-test-keypair"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-test"
  display_name = "terraform-test"
  service_offering= "Small Instance"
  network_id = "${cloudstack_network.foo.id}"
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = "Sandbox-simulator"
	keypair = "${cloudstack_ssh_keypair.foo.name}"
  expunge = true
}`

const testAccCloudStackInstance_project = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  project = "terraform"
  zone = "Sandbox-simulator"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-test"
  display_name = "terraform-test"
  service_offering= "Small Instance"
  network_id = "${cloudstack_network.foo.id}"
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  project = "terraform"
  zone = "${cloudstack_network.foo.zone}"
  expunge = true
}`
