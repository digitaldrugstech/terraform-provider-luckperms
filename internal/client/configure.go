package client

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// FromProviderData extracts *Client from Terraform provider data.
// Returns nil without error if data is nil (provider not yet configured).
func FromProviderData(data any, diags *diag.Diagnostics) *Client {
	if data == nil {
		return nil
	}
	c, ok := data.(*Client)
	if !ok {
		diags.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.Client, got %T", data),
		)
		return nil
	}
	return c
}
