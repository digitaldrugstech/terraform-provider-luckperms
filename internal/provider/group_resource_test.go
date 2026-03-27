package provider_test

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupResource_Basic(t *testing.T) {
	testAccPreCheck(t)
	name := randomName("acc_grp")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(name, "", 0, "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("luckperms_group.test", "name", name),
					resource.TestCheckResourceAttr("luckperms_group.test", "weight", "0"),
					resource.TestCheckResourceAttrSet("luckperms_group.test", "id"),
				),
			},
			{
				ResourceName:            "luckperms_group.test",
				ImportState:             true,
				ImportStateId:           name,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccGroupResource_WithMeta(t *testing.T) {
	testAccPreCheck(t)
	name := randomName("acc_meta")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(name, "Администрация", 500, `100.<#f1c40f>⭐`, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("luckperms_group.test", "display_name", "Администрация"),
					resource.TestCheckResourceAttr("luckperms_group.test", "weight", "500"),
					resource.TestCheckResourceAttr("luckperms_group.test", "prefix", "100.<#f1c40f>⭐"),
				),
			},
			{
				ResourceName:            "luckperms_group.test",
				ImportState:             true,
				ImportStateId:           name,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccGroupResource_UpdateMeta(t *testing.T) {
	testAccPreCheck(t)
	name := randomName("acc_upd")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(name, "OldName", 100, "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("luckperms_group.test", "display_name", "OldName"),
					resource.TestCheckResourceAttr("luckperms_group.test", "weight", "100"),
				),
			},
			{
				Config: testAccGroupConfig(name, "NewName", 200, "50.<red>★", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("luckperms_group.test", "display_name", "NewName"),
					resource.TestCheckResourceAttr("luckperms_group.test", "weight", "200"),
					resource.TestCheckResourceAttr("luckperms_group.test", "prefix", "50.<red>★"),
				),
			},
		},
	})
}

func TestAccGroupResource_DefaultGroup(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "luckperms_group" "test" {
  name = "default"
}
`,
				ImportState:   true,
				ResourceName:  "luckperms_group.test",
				ImportStateId: "default",
			},
		},
	})
}

func TestAccGroupResource_DefaultGroupDeleteNoop(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "luckperms_group" "default_test" {
  name = "default"
}
`,
				ImportState:   true,
				ResourceName:  "luckperms_group.default_test",
				ImportStateId: "default",
			},
			// Remove from config — destroy should be noop, default still exists
			{
				Config: `
data "luckperms_group" "verify_default" {
  name = "default"
}
`,
				Check: resource.TestCheckResourceAttr("data.luckperms_group.verify_default", "name", "default"),
			},
		},
	})
}

func TestAccGroupResource_ForceNew(t *testing.T) {
	testAccPreCheck(t)
	name1 := randomName("acc_rn1")
	name2 := randomName("acc_rn2")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(name1, "", 0, "", ""),
				Check:  resource.TestCheckResourceAttr("luckperms_group.test", "name", name1),
			},
			{
				Config: testAccGroupConfig(name2, "", 0, "", ""),
				Check:  resource.TestCheckResourceAttr("luckperms_group.test", "name", name2),
			},
		},
	})
}

func TestAccGroupResource_PreservesPermNodes(t *testing.T) {
	testAccPreCheck(t)
	name := randomName("acc_prsv")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "test" {
  name         = %q
  display_name = "Test"
  weight       = 10
}

resource "luckperms_group_nodes" "test" {
  group = luckperms_group.test.name

  node {
    key = "some.perm"
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("luckperms_group.test", "display_name", "Test"),
					resource.TestCheckResourceAttr("luckperms_group_nodes.test", "node.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "luckperms_group" "test" {
  name         = %q
  display_name = "UpdatedTest"
  weight       = 20
}

resource "luckperms_group_nodes" "test" {
  group = luckperms_group.test.name

  node {
    key = "some.perm"
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("luckperms_group.test", "display_name", "UpdatedTest"),
					resource.TestCheckResourceAttr("luckperms_group.test", "weight", "20"),
					resource.TestCheckResourceAttr("luckperms_group_nodes.test", "node.#", "1"),
				),
			},
		},
	})
}

func TestAccGroupResource_ConflictError(t *testing.T) {
	testAccPreCheck(t)
	// "default" always exists — creating it should fail with 409
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "luckperms_group" "conflict" {
  name = "default"
}
`,
				ExpectError: regexp.MustCompile(`(?s)already exists`),
			},
		},
	})
}

func TestAccGroupResource_DriftDetection(t *testing.T) {
	testAccPreCheck(t)
	name := randomName("acc_drift")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the group
			{
				Config: testAccGroupConfig(name, "DriftTest", 10, "", ""),
				Check:  resource.TestCheckResourceAttr("luckperms_group.test", "name", name),
			},
			// Delete externally, then plan — should detect removal
			{
				PreConfig: func() {
					baseURL := os.Getenv("LUCKPERMS_BASE_URL")
					req, _ := http.NewRequest("DELETE", baseURL+"/group/"+name, nil)
					if apiKey := os.Getenv("LUCKPERMS_API_KEY"); apiKey != "" {
						req.Header.Set("Authorization", "Bearer "+apiKey)
					}
					httpClient := &http.Client{Timeout: 10 * time.Second}
					resp, err := httpClient.Do(req)
					if err != nil {
						t.Fatalf("failed to delete group externally: %v", err)
					}
					resp.Body.Close()
					if resp.StatusCode != 200 {
						t.Fatalf("external delete returned %d, expected 200", resp.StatusCode)
					}
				},
				Config:             testAccGroupConfig(name, "DriftTest", 10, "", ""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGroupConfig(name, displayName string, weight int, prefix, suffix string) string {
	attrs := fmt.Sprintf(`  name = %q`, name)

	if displayName != "" {
		attrs += fmt.Sprintf("\n  display_name = %q", displayName)
	}
	if weight != 0 {
		attrs += fmt.Sprintf("\n  weight = %d", weight)
	}
	if prefix != "" {
		attrs += fmt.Sprintf("\n  prefix = %q", prefix)
	}
	if suffix != "" {
		attrs += fmt.Sprintf("\n  suffix = %q", suffix)
	}

	return fmt.Sprintf(`
resource "luckperms_group" "test" {
%s
}
`, attrs)
}
