package platform

import "testing"

func TestOperatorCredentialsMustBeConfigured(t *testing.T) {
	t.Setenv("SHIPPIN_OPERATOR_USER", "")
	t.Setenv("SHIPPIN_OPERATOR_PASS", "")
	if err := LoadConfig().ValidateOperatorCredentials(); err == nil {
		t.Fatal("empty environment must not supply an operator login")
	}
	for _, c := range []Config{
		{OperatorUser: "  ", OperatorPass: "example"},
		{OperatorUser: "example", OperatorPass: "  "},
		{OperatorUser: "example"},
		{OperatorPass: "example"},
	} {
		if err := c.ValidateOperatorCredentials(); err == nil {
			t.Error("missing or blank operator credentials accepted")
		}
	}
	t.Setenv("SHIPPIN_OPERATOR_USER", "local-operator")
	t.Setenv("SHIPPIN_OPERATOR_PASS", "local-test-value")
	if err := LoadConfig().ValidateOperatorCredentials(); err != nil {
		t.Fatalf("explicit credentials rejected: %v", err)
	}
}
