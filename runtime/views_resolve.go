package runtime

// views_resolve.go — the DSL value shapes for the control-plane
// recognition primitive resolve()/resolve_one() (api.md §5).
//
// resolve() returns a List<Match{ def, kind, score }>; resolve_one()
// returns the single best Match or Null. A Match is a read-only Getter
// view so routines read it as `m.def`, `m.kind`, `m.score`, plus the
// convenience hoists `m.id` / `m.canonical` / `m.source`.
//
// `m.def` re-uses the existing facts def-views (itemDefView,
// npcDefView, spellDefView, prayerDefView) so a resolved item reads
// exactly like an inventory slot's `.def`. Scenery/boundary ("loc")
// defs have no pre-existing view, so a small locDefView wraps them.

import (
	"github.com/gen0cide/westworld/cognition/resolve"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
)

// matchView wraps one resolve.Match for the DSL. It is immutable and
// carries no live state — a snapshot of one recognition result.
type matchView struct{ m resolve.Match }

func (v *matchView) Kind() string    { return "match" }
func (v *matchView) Display() string { return v.m.Canonical }

func (v *matchView) Get(field string) (interp.Value, bool) {
	switch field {
	case "def":
		// The concrete facts def, wrapped in the matching def-view so
		// the routine can read .id/.name/etc. Null if the def is
		// missing or an unrecognized type (defensive — should not
		// happen for a well-formed Match).
		if dv := defViewFor(v.m.Kind, v.m.Def); dv != nil {
			return dv, true
		}
		return interp.Null{}, true
	case "kind":
		return interp.String(v.m.Kind), true
	case "score":
		return interp.Float(v.m.Score), true
	case "id":
		return interp.Int(int64(v.m.ID)), true
	case "canonical", "name":
		return interp.String(v.m.Canonical), true
	case "source":
		// Which pipeline stage produced the match ("alias"/"fuzzy"/
		// "brain") — observability for routines that want to branch on
		// recognition confidence.
		return interp.String(v.m.Source), true
	}
	return nil, false
}

// defViewFor wraps a facts def (as carried by resolve.Match.Def) in the
// appropriate DSL def-view for its catalog kind. Returns nil when the
// def is nil or the (kind, concrete-type) pair is unrecognized, so the
// caller surfaces Null rather than a typed nil interface.
func defViewFor(kind string, def any) interp.Value {
	switch d := def.(type) {
	case *facts.ItemDef:
		if d != nil {
			return &itemDefView{def: d}
		}
	case *facts.NpcDef:
		if d != nil {
			return &npcDefView{def: d}
		}
	case *facts.SpellDef:
		if d != nil {
			return &spellDefView{def: d}
		}
	case *facts.PrayerDef:
		if d != nil {
			return &prayerDefView{def: d}
		}
	case *facts.SceneryDef:
		if d != nil {
			return &locDefView{scenery: d}
		}
	case *facts.BoundaryDef:
		if d != nil {
			return &locDefView{boundary: d}
		}
	}
	_ = kind // kind is implied by the concrete type; kept for callers/readers
	return nil
}

// locDefView wraps a scenery OR boundary def (the two facts catalogs
// that together make up resolve()'s "loc" kind) behind one read-only
// view. Exactly one of scenery/boundary is non-nil. Fields common to
// both (id, name, description, command1/command2) are exposed; the
// `is_boundary` flag lets a routine tell the two apart.
type locDefView struct {
	scenery  *facts.SceneryDef
	boundary *facts.BoundaryDef
}

func (d *locDefView) Kind() string { return "loc_def" }
func (d *locDefView) Display() string {
	if d.scenery != nil {
		return d.scenery.Name
	}
	if d.boundary != nil {
		return d.boundary.Name
	}
	return "<loc>"
}

func (d *locDefView) Get(field string) (interp.Value, bool) {
	switch field {
	case "id":
		if d.scenery != nil {
			return interp.Int(int64(d.scenery.ID)), true
		}
		return interp.Int(int64(d.boundary.ID)), true
	case "name":
		if d.scenery != nil {
			return interp.String(d.scenery.Name), true
		}
		return interp.String(d.boundary.Name), true
	case "description":
		if d.scenery != nil {
			return interp.String(d.scenery.Description), true
		}
		return interp.String(d.boundary.Description), true
	case "command1":
		if d.scenery != nil {
			return interp.String(d.scenery.Command1), true
		}
		return interp.String(d.boundary.Command1), true
	case "command2":
		if d.scenery != nil {
			return interp.String(d.scenery.Command2), true
		}
		return interp.String(d.boundary.Command2), true
	case "is_boundary":
		return interp.Bool(d.boundary != nil), true
	case "is_scenery":
		return interp.Bool(d.scenery != nil), true
	}
	return nil, false
}
