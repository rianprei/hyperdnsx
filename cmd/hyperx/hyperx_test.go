package main

import "testing"

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want map[string][]string
	}{
		{
			name: "empty",
			in:   "",
			want: nil,
		},
		{
			name: "single bearer",
			in:   "Authorization: Bearer abc123",
			want: map[string][]string{"Authorization": {"Bearer abc123"}},
		},
		{
			name: "multiple comma-separated",
			in:   "X-A: 1, X-B: 2",
			want: map[string][]string{"X-A": {"1"}, "X-B": {"2"}},
		},
		{
			name: "malformed pair skipped",
			in:   "no-colon, Good: yes",
			want: map[string][]string{"Good": {"yes"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHeaders(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("parseHeaders(%q) = %v, want %v", tt.in, got, tt.want)
			}
			for k, v := range tt.want {
				if len(got[k]) != len(v) || (len(v) > 0 && got[k][0] != v[0]) {
					t.Errorf("header %q = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestSplitComma(t *testing.T) {
	got := splitComma(" a , b ,, c ")
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("splitComma = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d = %q, want %q", i, got[i], want[i])
		}
	}
}
