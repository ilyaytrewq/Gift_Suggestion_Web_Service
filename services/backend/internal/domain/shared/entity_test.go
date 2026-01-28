package shared

import (
	"reflect"
	"testing"
)

func TestAgeLimitIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value AgeLimit
		want  bool
	}{
		{"none", AgeNone, true},
		{"age12", Age12, true},
		{"age16", Age16, true},
		{"age18", Age18, true},
		{"invalid", AgeLimit(5), false},
		{"negative", AgeLimit(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsValid(); got != tt.want {
				t.Fatalf("IsValid()=%v, want %v", got, tt.want)
			}
		})
	}
}

func TestMoneyIsNonNegative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value Money
		want  bool
	}{
		{"zero", Money(0), true},
		{"positive", Money(10), true},
		{"negative", Money(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsNonNegative(); got != tt.want {
				t.Fatalf("IsNonNegative()=%v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniqTags(t *testing.T) {
	t.Parallel()

	got := UniqTags([]TagID{"b", "a", "a"})
	want := []TagID{"a", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("UniqTags()=%v, want %v", got, want)
	}

	got = UniqTags(nil)
	if got != nil {
		t.Fatalf("UniqTags(nil)=%v, want nil", got)
	}
}
