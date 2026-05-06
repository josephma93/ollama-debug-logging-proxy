# Release Model

Homebrew release automation follows these rules:

- stable tags look like `v0.1.0`
- prerelease tags look like `v0.1.1-canary.1`, `v0.1.1-rc.1`, `v0.1.1-beta.1`, or `v0.1.1-alpha.1`
- any hyphen suffix after the numeric core marks a prerelease

Stable tags update the Homebrew formula. Prerelease tags publish release artifacts but do not update the stable Homebrew formula.

## Related

- [[0014-homebrew-tap-stable-canary-channels|0014 — Homebrew Tap with Stable + Canary Formulas]]
- [[home]]
