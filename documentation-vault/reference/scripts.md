# Scripts

Operational scripts live in `scripts/`:

- `install.sh`: source-checkout convenience path that builds the binary locally and then wires launchd
- `install-launchd.sh`: launchd setup path for an already-installed binary; intended to be idempotent and converge state
- `smoke-test.sh`: validates proxy health and basic tapped logging
- `uninstall-launchd.sh`: removes LaunchAgents and optionally proxy logs
- `uninstall.sh`: source-checkout convenience uninstall path

## Related

- [[install]]
- [[validate-setup]]
- [[launchd]]
- [[home]]
