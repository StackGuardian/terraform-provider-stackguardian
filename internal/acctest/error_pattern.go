package acctest

import (
	"regexp"
	"strings"
)

// TFStandardErrorPattern builds a regexp for resource.TestStep.ExpectError from a literal error
// message. Terraform's CLI hard-wraps diagnostic text at a fixed column width, so a long
// message can come back with newlines inserted between words (e.g. "TERRAFORM or\nOPENTOFU"
// instead of "TERRAFORM or OPENTOFU") — a pattern built with literal spaces silently fails
// to match whenever a wrap happens to land inside it. This quotes every word and joins them
// with \s+, so the match tolerates a line break (or any other run of whitespace) wherever
// the CLI happens to wrap.
func TFStandardErrorPattern(msg string) *regexp.Regexp {
	words := strings.Fields(msg)
	for i, w := range words {
		words[i] = regexp.QuoteMeta(w)
	}
	return regexp.MustCompile(strings.Join(words, `\s+`))
}
