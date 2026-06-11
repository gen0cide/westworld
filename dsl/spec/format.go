package spec

// format() template syntax — the single definition shared by the
// interpreter's renderer (dsl/interp/format.go) and the validator's
// static checks (dsl/validator). Keeping the scan here means the
// validator's placeholder count can never drift from what the
// renderer consumes.
//
// Syntax: only the exact pair {} is a placeholder (positional,
// consumed left-to-right). {{ renders a literal {, }} renders a
// literal }. Every other brace sequence is plain text — {name} is NOT
// special (str.format muscle memory; see FormatIdentPlaceholders for
// the validator warning that catches it).

// WalkFormatTemplate scans template once, firing callbacks in order:
// raw(s) for each run of plain literal text (contains no escapes or
// placeholders), escape(b) for each decoded {{ or }} (b is the
// literal brace), and placeholder() for each {}. Nil callbacks are
// skipped. Braces are ASCII, so the byte scan is UTF-8 safe.
func WalkFormatTemplate(template string, raw func(string), escape func(byte), placeholder func()) {
	flush := func(start, end int) {
		if raw != nil && start < end {
			raw(template[start:end])
		}
	}
	start := 0
	for i := 0; i < len(template); {
		c := template[i]
		if c == '{' && i+1 < len(template) && template[i+1] == '{' {
			flush(start, i)
			if escape != nil {
				escape('{')
			}
			i += 2
			start = i
			continue
		}
		if c == '{' && i+1 < len(template) && template[i+1] == '}' {
			flush(start, i)
			if placeholder != nil {
				placeholder()
			}
			i += 2
			start = i
			continue
		}
		if c == '}' && i+1 < len(template) && template[i+1] == '}' {
			flush(start, i)
			if escape != nil {
				escape('}')
			}
			i += 2
			start = i
			continue
		}
		i++
	}
	flush(start, len(template))
}

// CountFormatPlaceholders returns the number of {} placeholders in
// template. Escapes ({{ / }}) don't count.
func CountFormatPlaceholders(template string) int {
	n := 0
	WalkFormatTemplate(template, nil, nil, func() { n++ })
	return n
}

// FormatIdentPlaceholders returns each identifier that appears as
// {ident} in template's RAW literal text — almost certainly
// str.format muscle memory, since the {} contract renders it
// literally. Escaped braces don't participate ({{name}} was written
// deliberately), and {0}-style indexed forms are not idents. Order of
// appearance, duplicates preserved.
//
// The closing brace of an {ident} can be claimed by a }} escape — in
// "{name}}" the walker's raw run ends at "{name" — so a run ending in
// "{ident" whose NEXT event is an escaped '}' still counts ("{name}}"
// and "{{{name}}}" are exactly the muscle-memory shapes this exists
// to catch).
func FormatIdentPlaceholders(template string) []string {
	var out []string
	pending := "" // raw run ended in "{ident"; resolved by the next event
	WalkFormatTemplate(template, func(s string) {
		pending = ""
		for i := 0; i < len(s); i++ {
			if s[i] != '{' || i+1 >= len(s) || !isFormatIdentStart(s[i+1]) {
				continue
			}
			j := i + 1
			for j < len(s) && isFormatIdentChar(s[j]) {
				j++
			}
			if j == len(s) {
				pending = s[i+1:]
			} else if s[j] == '}' {
				out = append(out, s[i+1:j])
				i = j
			}
		}
	}, func(b byte) {
		if b == '}' && pending != "" {
			out = append(out, pending)
		}
		pending = ""
	}, func() {
		pending = ""
	})
	return out
}

func isFormatIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isFormatIdentChar(c byte) bool {
	return isFormatIdentStart(c) || (c >= '0' && c <= '9')
}
