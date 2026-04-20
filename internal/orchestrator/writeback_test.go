package orchestrator

import "testing"

func TestQueryWriteBackDocID_stable(t *testing.T) {
	a := queryWriteBackDocID("ase-q-", "  hello  ")
	b := queryWriteBackDocID("ase-q-", "hello")
	if a != b {
		t.Fatalf("trim mismatch: %q vs %q", a, b)
	}
	if len(a) < len("ase-q-")+32 {
		t.Fatalf("unexpected id len: %q", a)
	}
}
