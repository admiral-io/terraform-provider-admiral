package provider

import "testing"

func TestFilterEq(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "plain", value: "my-app", want: "field['name'] = 'my-app'"},
		{name: "embedded quote is escaped", value: "o'brien", want: `field['name'] = 'o\'brien'`},
		{name: "quote breakout is neutralised", value: "x' OR field['name'] = 'y", want: `field['name'] = 'x\' OR field[\'name\'] = \'y'`},
		{name: "trailing backslash is rejected", value: `bad\`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := filterEq("name", tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("filterEq(%q) = %q, want error", tt.value, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("filterEq(%q) unexpected error: %v", tt.value, err)
			}
			if got != tt.want {
				t.Fatalf("filterEq(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
