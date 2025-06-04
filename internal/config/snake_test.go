package config

import "testing"

func TestToSnakeCase(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"TestCamelCase", "test_camel_case"},
		{"JWKSPrivate", "jwks_private"},
		{"HTTPServerURL", "http_server_url"},
		{"UserID", "user_id"},
		{"API", "api"},
	}

	for _, c := range cases {
		got := toSnakeCase(c.in)
		if got != c.want {
			t.Errorf("toSnakeCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
