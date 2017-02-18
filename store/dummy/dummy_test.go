package dummy

import (
	"testing"

	"github.com/jasonkeene/anubot-server/store"
)

func TestThatDummyBackendCompliesWithAllStoreMethods(t *testing.T) {
	var _ store.Store = &Dummy{}
}
