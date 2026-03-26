package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupNodesResource_Basic(t *testing.T) {
	testAccPreCheck(t)
	name := randomName("acc_nd")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "test" {
  name = %q
}

resource "luckperms_group_nodes" "test" {
  group = luckperms_group.test.name

  node {
    key   = "test.permission.one"
    value = true
  }

  node {
    key   = "test.permission.two"
    value = false
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("luckperms_group_nodes.test", "group", name),
					resource.TestCheckResourceAttr("luckperms_group_nodes.test", "id", name),
					resource.TestCheckResourceAttr("luckperms_group_nodes.test", "node.#", "2"),
				),
			},
			{
				ResourceName:                         "luckperms_group_nodes.test",
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
		},
	})
}

func TestAccGroupNodesResource_Update(t *testing.T) {
	testAccPreCheck(t)
	name := randomName("acc_ndupd")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "test" {
  name = %q
}

resource "luckperms_group_nodes" "test" {
  group = luckperms_group.test.name

  node {
    key = "perm.a"
  }
}
`, name),
				Check: resource.TestCheckResourceAttr("luckperms_group_nodes.test", "node.#", "1"),
			},
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "test" {
  name = %q
}

resource "luckperms_group_nodes" "test" {
  group = luckperms_group.test.name

  node { key = "perm.a" }
  node { key = "perm.b" }
  node {
    key   = "perm.c"
    value = false
  }
}
`, name),
				Check: resource.TestCheckResourceAttr("luckperms_group_nodes.test", "node.#", "3"),
			},
		},
	})
}

func TestAccGroupNodesResource_WithContext(t *testing.T) {
	testAccPreCheck(t)
	name := randomName("acc_ndctx")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "test" {
  name = %q
}

resource "luckperms_group_nodes" "test" {
  group = luckperms_group.test.name

  node {
    key = "worldedit.*"
    context {
      key   = "server"
      value = "creative-build"
    }
    context {
      key   = "server"
      value = "creative-infrastructure"
    }
  }

  node {
    key = "plots.admin"
    context {
      key   = "server"
      value = "creative-build"
    }
  }

  node {
    key = "global.perm"
  }
}
`, name),
				Check: resource.TestCheckResourceAttr("luckperms_group_nodes.test", "node.#", "3"),
			},
			{
				ResourceName:                         "luckperms_group_nodes.test",
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
		},
	})
}

func TestAccGroupNodesResource_Inheritance(t *testing.T) {
	testAccPreCheck(t)
	parent := randomName("acc_par")
	child := randomName("acc_ch")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "parent" {
  name = %q
}

resource "luckperms_group" "child" {
  name = %q
}

resource "luckperms_group_nodes" "child" {
  group = luckperms_group.child.name

  node {
    key = "group.%s"
  }

  node {
    key = "some.perm"
  }
}
`, parent, child, parent),
				Check: resource.TestCheckResourceAttr("luckperms_group_nodes.child", "node.#", "2"),
			},
		},
	})
}

func TestAccGroupNodesResource_NegatedPermissions(t *testing.T) {
	testAccPreCheck(t)
	name := randomName("acc_ndneg")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "test" {
  name = %q
}

resource "luckperms_group_nodes" "test" {
  group = luckperms_group.test.name

  node {
    key   = "*"
    value = true
  }

  node {
    key   = "vulcan.bypass.*"
    value = false
  }
}
`, name),
				Check: resource.TestCheckResourceAttr("luckperms_group_nodes.test", "node.#", "2"),
			},
			{
				ResourceName:                         "luckperms_group_nodes.test",
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
		},
	})
}

func TestAccGroupNodesResource_MetaNodeValidation(t *testing.T) {
	testAccPreCheck(t)
	name := randomName("acc_mval")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "test" {
  name = %q
}

resource "luckperms_group_nodes" "test" {
  group = luckperms_group.test.name
  node { key = "displayname.ShouldFail" }
}
`, name),
				ExpectError: regexp.MustCompile(`(?s)meta node`),
			},
		},
	})
}

func TestAccGroupNodesResource_MetaNodeValidation_Weight(t *testing.T) {
	testAccPreCheck(t)
	name := randomName("acc_mval2")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "test" {
  name = %q
}

resource "luckperms_group_nodes" "test" {
  group = luckperms_group.test.name
  node { key = "weight.500" }
}
`, name),
				ExpectError: regexp.MustCompile(`(?s)meta node`),
			},
		},
	})
}

func TestAccGroupNodesResource_PreservesMetaOnDelete(t *testing.T) {
	testAccPreCheck(t)
	name := randomName("acc_pmd")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "test" {
  name         = %q
  display_name = "KeepThis"
  weight       = 42
}

resource "luckperms_group_nodes" "test" {
  group = luckperms_group.test.name
  node { key = "some.perm" }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("luckperms_group.test", "display_name", "KeepThis"),
					resource.TestCheckResourceAttr("luckperms_group_nodes.test", "node.#", "1"),
				),
			},
			// Remove group_nodes — meta nodes must survive
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "test" {
  name         = %q
  display_name = "KeepThis"
  weight       = 42
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("luckperms_group.test", "display_name", "KeepThis"),
					resource.TestCheckResourceAttr("luckperms_group.test", "weight", "42"),
				),
			},
		},
	})
}
