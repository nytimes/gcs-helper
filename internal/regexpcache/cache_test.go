package regexpcache

import (
	"os"
	"testing"
)

func TestMustCompile(t *testing.T) {
	t.Parallel()

	re := MustCompile("^$")
	cached := MustCompile("^$")
	if re != cached {
		t.Errorf("didn't cache regexp\nwant %#v\ngot  %#v", re, cached)
	}

	other := MustCompile("^ $")
	if other == re {
		t.Errorf("polluted cache?")
	}
}

func TestMustCompilePanicsOnInvalidRegexp(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("didn't panic on invalid regexp")
		}
	}()

	MustCompile("[0----]")
}

func TestMain(m *testing.M) {
	cache.Range(func(k, _ interface{}) bool {
		cache.Delete(k)
		return true
	})
	os.Exit(m.Run())
}
