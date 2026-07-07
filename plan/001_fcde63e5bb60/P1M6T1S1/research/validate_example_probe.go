//go:build ignore
// +build ignore
//
// Research probe (P1.M6.T1.S1): validates the example skill frontmatter against
// PRD §9 rules WITHOUT depending on `skpp check` (M4.T10 may not be landed yet).
// Run from repo root:  go run plan/.../research/validate_example_probe.go <SKILL.md>
package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/dabstractor/skpp/internal/discover"
)

func main() {
	path := os.Args[1]
	fm, body, err := discover.ParseFrontmatter(path)
	if err != nil { fmt.Println("PARSE ERROR:", err); os.Exit(1) }
	if !fm.HasFM { fmt.Println("FAIL: no frontmatter block"); os.Exit(1) }
	nameRe := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	if fm.Name == "" { fmt.Println("FAIL: name empty"); os.Exit(1) }
	if !nameRe.MatchString(fm.Name) || len(fm.Name) > 64 { fmt.Println("FAIL: name invalid:", fm.Name); os.Exit(1) }
	if fm.Description == "" { fmt.Println("FAIL: description empty"); os.Exit(1) }
	if len(fm.Description) > 1024 { fmt.Println("WARN: description >1024") }
	s := discover.BuildSkill("/skills/example", "example", fm)
	fmt.Printf("OK   example (%s)\n", s.Name)
	fmt.Printf("  HasFM=%v keywords=%v category=%q aliases=%v\n", s.HasFM, s.Keywords, s.Category, s.Aliases)
	fmt.Printf("  description chars=%d body_present=%v\n", len(s.Description), body != "")
}
