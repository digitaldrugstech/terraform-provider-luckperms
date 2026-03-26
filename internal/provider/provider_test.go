package provider_test

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"luckperms": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("LUCKPERMS_BASE_URL") == "" {
		// Default matches docker-compose.yml (9094:8080). CI overrides via env var.
		os.Setenv("LUCKPERMS_BASE_URL", "http://localhost:9094")
	}
}

func randomName(prefix string) string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%s_%x", prefix, b)
}
