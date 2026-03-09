package tenancy

import "testing"

func TestAccountCreateValidate(t *testing.T) {
	create, err := (AccountCreate{Name: "  Acme  "}).Validate()
	if err != nil {
		t.Fatalf("validate create: %v", err)
	}
	if create.Name != "Acme" {
		t.Fatalf("expected trimmed name, got %q", create.Name)
	}
	if _, err := (AccountCreate{Name: "   "}).Validate(); err == nil {
		t.Fatal("expected validation error for blank name")
	}
}
