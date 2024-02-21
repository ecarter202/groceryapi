package shared

import "testing"

func TestGenProductCode(t *testing.T) {
	wantedLength := 17
	code := GenProductCode()
	if len(code) != wantedLength {
		t.Errorf("wanted product code length of %d but is of length %d", wantedLength, len(code))
	}
}
