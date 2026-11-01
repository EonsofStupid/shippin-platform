package identity

import "testing"

func TestUnconfiguredOperatorCannotLogIn(t *testing.T) {
	for _, credentials := range [][2]string{{"", ""}, {" ", "example"}, {"example", " "}} {
		store := NewStore(credentials[0], credentials[1])
		if token, _, err := store.Login(credentials[0], credentials[1]); err == nil || token != "" {
			t.Fatal("unconfigured operator issued a session")
		}
	}
}
