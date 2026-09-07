package acctest

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
)

const (
	// charSetAlphaNum is the alphanumeric character set for use with
	// RandStringFromCharSet.
	charSetAlphaNum = "abcdefghijklmnopqrstuvwxyz012346789"

	// charSetAlpha is the alphabetical character set for use with
	// RandStringFromCharSet.
	charSetAlpha = "abcdefghijklmnopqrstuvwxyz"

	// Length of the resource name we wish to generate.
	resourceNameLength = 10

	// suffixLength is the random tail appended by ResourceName. Eight characters
	// of lowercase alpha is ~200 billion combinations, which is far more than a
	// test suite needs to avoid colliding with a leftover from an earlier run.
	suffixLength = 8

	// NamePrefix marks every identifier these tests create. Randomising a name
	// stops a failed cleanup from blocking the next run, but on its own it turns
	// each leak into an unattributable resource nobody dares delete. The shared
	// prefix is what makes leaks identifiable afterwards -- by the sweeper, by a
	// dashboard filter, or by a human deciding what is safe to remove.
	NamePrefix = "tfacc"
)

// ResourceName returns a unique identifier for one acceptance-test resource,
// shaped as "tfacc-<slug>-<random>".
//
// Static names made the suite order-dependent: a run that failed before its
// destroy step left the name taken, and every later run failed at create with
// 409 "already exists" until somebody deleted it by hand. A fresh name per run
// removes that coupling.
//
// The slug is kept so a leaked resource still says which test made it.
func ResourceName(slug string) string {
	return fmt.Sprintf("%s-%s-%s", NamePrefix, sanitizeSlug(slug), randomSuffix())
}

// ResourceNameRaw makes a name unique without cleaning it up, for the tests whose
// subject is an awkward name: one carrying spaces, capitals, or punctuation the
// platform has to cope with. Sanitising those would leave the test passing while
// removing the thing it checks.
//
// The prefix and random part lead so the result is still recognisable to
// IsTestResourceName and to a sweeper.
func ResourceNameRaw(name string) string {
	return fmt.Sprintf("%s-%s %s", NamePrefix, randomSuffix(), name)
}

// IsTestResourceName reports whether a name looks like one this suite created.
// Sweepers use it to decide what is safe to delete, so it is deliberately strict:
// anything not carrying the prefix is left alone.
func IsTestResourceName(name string) bool {
	return strings.HasPrefix(name, NamePrefix+"-")
}

// sanitizeSlug keeps the descriptive part usable as an identifier: an existing
// name may contain spaces or capitals, neither of which belongs in an ID, and an
// over-long slug pushes the name past the platform's length limits.
func sanitizeSlug(slug string) string {
	slug = strings.ToLower(slug)
	var b strings.Builder
	for _, r := range slug {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-', r == '_', r == ' ':
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	if len(out) > 40 {
		out = strings.Trim(out[:40], "-")
	}
	if out == "" {
		out = "resource"
	}
	return out
}

func randomSuffix() string {
	result := make([]byte, suffixLength)
	for i := range result {
		result[i] = charSetAlpha[randIntRange(0, len(charSetAlpha))]
	}
	return string(result)
}

// GenerateRandomResourceName builds a unique-ish resource identifier to use in
// tests.
//
// Deprecated: prefer ResourceName, whose output carries the shared prefix and so
// can be recognised and swept up after a failed run.
func GenerateRandomResourceName() string {
	result := make([]byte, resourceNameLength)
	for i := 0; i < resourceNameLength; i++ {
		result[i] = charSetAlpha[randIntRange(0, len(charSetAlpha))]
	}
	return string(result)
}

// randIntRange returns a random integer between min (inclusive) and max
// (exclusive).
func randIntRange(min int, max int) int {
	return rand.Intn(max-min) + min
}

// SweepEnabled reports whether the caller asked for a cleanup pass rather than a
// normal test run. Kept separate from the sweeper implementations so a package
// can guard destructive helpers behind it.
func SweepEnabled() bool {
	return os.Getenv("SG_ACC_SWEEP") != ""
}
