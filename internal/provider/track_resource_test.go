package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTrackResource_Basic(t *testing.T) {
	testAccPreCheck(t)
	grpA := randomName("acc_trga")
	grpB := randomName("acc_trgb")
	track := randomName("acc_trk")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "a" { name = %q }
resource "luckperms_group" "b" { name = %q }

resource "luckperms_track" "test" {
  name   = %q
  groups = [luckperms_group.a.name, luckperms_group.b.name]
}
`, grpA, grpB, track),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("luckperms_track.test", "name", track),
					resource.TestCheckResourceAttr("luckperms_track.test", "groups.#", "2"),
					resource.TestCheckResourceAttr("luckperms_track.test", "groups.0", grpA),
					resource.TestCheckResourceAttr("luckperms_track.test", "groups.1", grpB),
					resource.TestCheckResourceAttrSet("luckperms_track.test", "id"),
				),
			},
			{
				ResourceName:            "luckperms_track.test",
				ImportState:             true,
				ImportStateId:           track,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccTrackResource_Update(t *testing.T) {
	testAccPreCheck(t)
	g1 := randomName("acc_tu1")
	g2 := randomName("acc_tu2")
	track := randomName("acc_tru")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "g1" { name = %q }
resource "luckperms_group" "g2" { name = %q }

resource "luckperms_track" "test" {
  name   = %q
  groups = [luckperms_group.g1.name]
}
`, g1, g2, track),
				Check: resource.TestCheckResourceAttr("luckperms_track.test", "groups.#", "1"),
			},
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "g1" { name = %q }
resource "luckperms_group" "g2" { name = %q }

resource "luckperms_track" "test" {
  name   = %q
  groups = [luckperms_group.g1.name, luckperms_group.g2.name]
}
`, g1, g2, track),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("luckperms_track.test", "groups.#", "2"),
					resource.TestCheckResourceAttr("luckperms_track.test", "groups.0", g1),
					resource.TestCheckResourceAttr("luckperms_track.test", "groups.1", g2),
				),
			},
		},
	})
}

func TestAccTrackResource_EmptyGroups(t *testing.T) {
	testAccPreCheck(t)
	track := randomName("acc_tre")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_track" "test" {
  name   = %q
  groups = []
}
`, track),
				Check: resource.TestCheckResourceAttr("luckperms_track.test", "groups.#", "0"),
			},
		},
	})
}

func TestAccTrackResource_GroupOrder(t *testing.T) {
	testAccPreCheck(t)
	gc := randomName("acc_toc")
	ga := randomName("acc_toa")
	gb := randomName("acc_tob")
	track := randomName("acc_tro")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "c" { name = %q }
resource "luckperms_group" "a" { name = %q }
resource "luckperms_group" "b" { name = %q }

resource "luckperms_track" "test" {
  name   = %q
  groups = [luckperms_group.c.name, luckperms_group.a.name, luckperms_group.b.name]
}
`, gc, ga, gb, track),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("luckperms_track.test", "groups.0", gc),
					resource.TestCheckResourceAttr("luckperms_track.test", "groups.1", ga),
					resource.TestCheckResourceAttr("luckperms_track.test", "groups.2", gb),
				),
			},
		},
	})
}
