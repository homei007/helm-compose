package provider

import (
	"reflect"
	"testing"
)

func TestRevisionsToDeleteKeepsExactRetentionCount(t *testing.T) {
	if got, want := revisionsToDelete(1, 6, 3), []int{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected expired revisions: got %#v, want %#v", got, want)
	}
	if got := revisionsToDelete(4, 6, 3); len(got) != 0 {
		t.Fatalf("expected no expired revisions, got %#v", got)
	}
}
