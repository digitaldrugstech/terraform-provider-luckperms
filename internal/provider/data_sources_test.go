package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupDataSource(t *testing.T) {
	testAccPreCheck(t)
	name := randomName("acc_dsg")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "test" {
  name         = %q
  display_name = "TestDisplay"
  weight       = 42
  prefix       = "10.<green>T"
}

resource "luckperms_group_nodes" "test" {
  group = luckperms_group.test.name
  node { key = "some.perm" }
}

data "luckperms_group" "test" {
  name       = luckperms_group.test.name
  depends_on = [luckperms_group_nodes.test]
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.luckperms_group.test", "name", name),
					resource.TestCheckResourceAttr("data.luckperms_group.test", "weight", "42"),
					resource.TestCheckResourceAttr("data.luckperms_group.test", "display_name", "TestDisplay"),
					resource.TestCheckResourceAttr("data.luckperms_group.test", "prefix", "10.<green>T"),
					resource.TestCheckResourceAttr("data.luckperms_group.test", "nodes.#", "1"),
				),
			},
		},
	})
}

func TestAccGroupsDataSource(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "luckperms_groups" "all" {}`,
				Check:  resource.TestCheckResourceAttrSet("data.luckperms_groups.all", "names.#"),
			},
		},
	})
}

func TestAccTrackDataSource(t *testing.T) {
	testAccPreCheck(t)
	grp := randomName("acc_dst")
	track := randomName("acc_dstrk")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "test" { name = %q }

resource "luckperms_track" "test" {
  name   = %q
  groups = [luckperms_group.test.name]
}

data "luckperms_track" "test" {
  name = luckperms_track.test.name
}
`, grp, track),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.luckperms_track.test", "name", track),
					resource.TestCheckResourceAttr("data.luckperms_track.test", "groups.#", "1"),
					resource.TestCheckResourceAttr("data.luckperms_track.test", "groups.0", grp),
				),
			},
		},
	})
}

func TestAccTracksDataSource(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "luckperms_tracks" "all" {}`,
				Check:  resource.TestCheckResourceAttrSet("data.luckperms_tracks.all", "names.#"),
			},
		},
	})
}
