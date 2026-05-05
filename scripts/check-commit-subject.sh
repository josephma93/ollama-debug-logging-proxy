#!/usr/bin/env sh
# Validate a commit subject against the project's Conventional Commit format.
# Used by .githooks/commit-msg (locally) and .github/workflows/ci.yml (in CI).

set -eu

subject="${1-}"

if [ -z "$subject" ]; then
	exit 0
fi

case "$subject" in
	"Revert "*) exit 0 ;;
esac

pattern='^(feat|fix|chore|docs|refactor|test|ci|build|perf|revert|style|deps)(\([^)]+\))?!?: .+'

if printf '%s' "$subject" | grep -Eq "$pattern"; then
	exit 0
fi

cat >&2 <<EOF
Commit subject does not match Conventional Commit format.

Got:
  $subject

Expected:
  <type>(<optional scope>)<optional !>: <imperative summary>

Allowed types: feat, fix, chore, docs, refactor, test, ci, build, perf,
               revert, style, deps

Examples:
  feat: add tail subcommand
  fix(proxy): handle empty body
  chore!: drop legacy env var
EOF
exit 1
