package dhcp6

import (
	"testing"
)



func TestUnmarshalFailsIfOROLengthIsOdd(t *testing.T) {
	in := []byte{0, 6, 0, 3, 0, 1, 1}
	if _, err := MakeOptions(in); err == nil {
		t.Fatalf("Parsing options should fail: option request for options has odd length.")
	}
}

func TestUnmarshalORO(t *testing.T) {
	options := Options{
		6: [][]byte{{0, 1, 0, 5, 1, 0}},
	}
	parsed_options := options.UnmarshalOptionRequestOption()

	if l := len(parsed_options); l != 3 {
		t.Fatalf("Expected 3 options, got: %d", l)
	}

	if _, present := parsed_options[1]; !present {
		t.Fatalf("Should contain option id 1")
	}

	if _, present := parsed_options[5]; !present {
		t.Fatalf("Should contain option id 5")
	}

	if _, present := parsed_options[256]; !present {
		t.Fatalf("Should contain option id 256")
	}
}
