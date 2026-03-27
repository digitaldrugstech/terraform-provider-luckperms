package provider_test

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

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

func TestAccTrackResource_ConflictError(t *testing.T) {
	testAccPreCheck(t)
	track := randomName("acc_trc")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Pre-create the track via HTTP so Terraform hits a conflict on Create.
				PreConfig: func() {
					baseURL := os.Getenv("LUCKPERMS_BASE_URL")
					body := fmt.Sprintf(`{"name":%q}`, track)
					req, _ := http.NewRequest(http.MethodPost, baseURL+"/track", strings.NewReader(body))
					req.Header.Set("Content-Type", "application/json")
					if apiKey := os.Getenv("LUCKPERMS_API_KEY"); apiKey != "" {
						req.Header.Set("Authorization", "Bearer "+apiKey)
					}
					httpClient := &http.Client{Timeout: 10 * time.Second}
					resp, err := httpClient.Do(req)
					if err != nil {
						t.Fatalf("pre-create track failed: %v", err)
					}
					resp.Body.Close()
				},
				Config: fmt.Sprintf(`
resource "luckperms_track" "conflict" {
  name   = %q
  groups = []
}
`, track),
				ExpectError: regexp.MustCompile(`(?s)already exists`),
			},
		},
	})
}

func TestAccTrackResource_DriftDetection(t *testing.T) {
	testAccPreCheck(t)
	track := randomName("acc_trd")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the track via Terraform.
			{
				Config: fmt.Sprintf(`
resource "luckperms_track" "test" {
  name   = %q
  groups = []
}
`, track),
				Check: resource.TestCheckResourceAttr("luckperms_track.test", "name", track),
			},
			// Delete externally, then plan — should detect drift.
			{
				PreConfig: func() {
					baseURL := os.Getenv("LUCKPERMS_BASE_URL")
					req, _ := http.NewRequest(http.MethodDelete, baseURL+"/track/"+track, nil)
					if apiKey := os.Getenv("LUCKPERMS_API_KEY"); apiKey != "" {
						req.Header.Set("Authorization", "Bearer "+apiKey)
					}
					httpClient := &http.Client{Timeout: 10 * time.Second}
					resp, err := httpClient.Do(req)
					if err != nil {
						t.Fatalf("failed to delete track externally: %v", err)
					}
					resp.Body.Close()
					if resp.StatusCode != 200 {
						t.Fatalf("external delete returned %d, expected 200", resp.StatusCode)
					}
				},
				Config: fmt.Sprintf(`
resource "luckperms_track" "test" {
  name   = %q
  groups = []
}
`, track),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
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
