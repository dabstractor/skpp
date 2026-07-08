---
name: example
description: >
  Reference example skill for skilldozer. Demonstrates the required frontmatter and
  how skilldozer resolves a tag to an absolute path. Safe to delete once you add real skills.
metadata:
  keywords: [example, demo, skilldozer]
  category: meta
---

# Example Skill

This skill exists only so `skilldozer` has something to resolve.

Try:

```bash
skilldozer example                       # prints this directory's absolute path
skilldozer -f example                    # prints .../skills/example/SKILL.md
pi --skill "$(skilldozer example)"       # loads this skill into pi
```
