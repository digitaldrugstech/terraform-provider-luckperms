package client

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestFromProviderData_Nil(t *testing.T) {
	var diags diag.Diagnostics
	c := FromProviderData(nil, &diags)
	if c != nil {
		t.Errorf("expected nil client for nil data, got %v", c)
	}
	if diags.HasError() {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

func TestFromProviderData_WrongType(t *testing.T) {
	var diags diag.Diagnostics
	c := FromProviderData("not-a-client", &diags)
	if c != nil {
		t.Errorf("expected nil client for wrong type, got %v", c)
	}
	if !diags.HasError() {
		t.Error("expected diagnostic error for wrong type")
	}
}

func TestFromProviderData_Valid(t *testing.T) {
	expected := &Client{BaseURL: "http://localhost:8080", APIKey: "key"}
	var diags diag.Diagnostics
	c := FromProviderData(expected, &diags)
	if c != expected {
		t.Errorf("expected returned client to be the same pointer")
	}
	if diags.HasError() {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}
