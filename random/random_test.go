package random

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	opts := Options{Length: 5}
	got, err := Generate(opts)
	if err != nil {
		t.Errorf("Generate() = %v, failed, check fn", err)
		return
	}
	if len(got) != opts.Length {
		t.Errorf("wanted string of length %d, got string of length %d", opts.Length, len(got))
	}
}
