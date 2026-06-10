package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/parser"
)

// TestDetourIntentSourcesParse guards the embedded reflex routines: a parse
// error here means the detour fires and immediately dies, silently breaking
// the reflex (survival eat / fatigue sleep).
func TestDetourIntentSourcesParse(t *testing.T) {
	for _, in := range []Intent{survivalIntent(), fatigueIntent()} {
		if _, err := parser.Parse(in.Name, in.Source); err != nil {
			t.Errorf("%s source does not parse: %v", in.Name, err)
		}
	}
}
