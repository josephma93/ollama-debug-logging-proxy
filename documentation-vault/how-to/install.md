# Install

`ollama-logging-proxy-install` is the command that rewires the local Ollama setup. It:

- installs or updates two user LaunchAgents in `~/Library/LaunchAgents`
- changes Ollama from public `11434` serving to private `127.0.0.1:11435`
- starts the proxy in front of Ollama on `11434`
- preserves existing Ollama `OLLAMA_DEBUG`, `OLLAMA_MODELS`, `HOME`, and `PATH` values across re-runs unless explicitly overridden

This is intentionally a user-scope `launchd` setup, not a system daemon install.

What it does not do:

- it does not use `brew services`
- it does not install anything into `/Library/LaunchDaemons`
- it does not keep Ollama directly exposed on `11434`
- it does not change non-macOS service managers
- it does not promise cross-platform deployment behavior

## Related

- [[validate-setup]]
- [[scripts]]
- [[launchd]]
- [[deployment-topology]]
- [[0003-user-launchagent-scope|0003 — macOS User LaunchAgent]]
- [[0014-homebrew-tap-stable-canary-channels|0014 — Homebrew Tap with Stable + Canary Formulas]]
- [[home]]
