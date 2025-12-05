# eightctl

A modern Go CLI for Eight Sleep Pods. Control power/temperature, alarms, audio, base, autopilot, travel, household, and export sleep metrics. Includes a daemon for scheduled routines.

> Eight Sleep does **not** publish a stable public API. `eightctl` speaks the same undocumented cloud endpoints the mobile apps use. Default OAuth client creds are baked in (from Android APK 7.39.17), so you generally only provide email + password.

## Quickstart
```bash
# build + install
GO111MODULE=on go install github.com/steipete/eightctl/cmd/eightctl@latest

# create config (optional; flags/env also work)
mkdir -p ~/.config/eightctl
cat > ~/.config/eightctl/config.yaml <<'CFG'
email: "you@example.com"
password: "your-password"
# user_id: "optional"               # auto-resolved via /users/me
# timezone: "America/New_York"      # defaults to local
# client_id / client_secret optional # defaults to app creds
CFG
chmod 600 ~/.config/eightctl/config.yaml

# check pod state
EIGHTCTL_EMAIL=you@example.com EIGHTCTL_PASSWORD=your-password eightctl status

# set temperature level (-100..100)
eightctl temp 20

# run daemon with your YAML schedule (see docs/example-schedule.yaml)
eightctl daemon --dry-run
```

## Commands (high level)
- **Power & temp:** `on`, `off`, `temp <level>`, `status`
- **Schedules & daemon:** `schedule list|create|update|delete|next`, `daemon`
- **Alarms:** `alarm list|create|update|delete|snooze|dismiss|dismiss-all|vibration-test`
- **Temperature modes:** `tempmode nap on|off|extend|status`, `tempmode hotflash on|off|status`, `tempmode events`
- **Audio:** `audio tracks|categories|state|play|pause|seek|volume|pair|next`, `audio favorites list|add|remove`
- **Base:** `base info|angle|presets|preset-run|vibration-test`
- **Device:** `device info|peripherals|owner|warranty|online|priming-tasks|priming-schedule`
- **Metrics & insights:** `sleep day|range`, `presence`, `metrics trends|intervals|summary|aggregate|insights`
- **Autopilot:** `autopilot details|history|recap`, `autopilot set-level-suggestions`, `autopilot set-snore-mitigation`
- **Travel:** `travel trips|create-trip|delete-trip|plans|create-plan|update-plan|tasks|airport-search|flight-status`
- **Household:** `household summary|schedule|current-set|invitations|devices|users|guests`
- **Misc:** `tracks`, `feats`, `whoami`, `version`

Use `--output table|json|csv` and `--fields field1,field2` to shape output. `--verbose` enables debug logs; `--quiet` hides the config banner.

## Configuration
Priority: flags > env vars (`EIGHTCTL_*`) > config file.

Key fields: `email`, `password`, optional `user_id`, `client_id`, `client_secret`, `timezone`, `output`, `fields`, `verbose`. The client auto-resolves `user_id` and `device_id` after authentication.

## Tests / Lint / Format
```bash
make fmt      # gofumpt -w ./
make lint     # golangci-lint run ./...
make test     # go test ./...
```
CI runs the same (see .github/workflows/ci.yml).

### pnpm scripts
If you prefer `pnpm`, a minimal `package.json` is included:

```bash
pnpm eightctl   # go run ./cmd/eightctl
pnpm start      # same as above
pnpm build      # go build -o bin/eightctl
pnpm lint       # golangci-lint run ./...
pnpm format     # gofumpt -w .
pnpm test       # go test ./...
```

Tools such as `gofumpt` and `golangci-lint` must be installed in your PATH (see CI for versions).

## Security
- Config permissions are checked; we warn if looser than `0600`.
- Credentials are only sent to Eight Sleep over HTTPS; they are not persisted elsewhere by the CLI.

## Prior Work / References
- Go CLI `clim8`: https://github.com/blacktop/clim8
- MCP server (Node/TS): https://github.com/elizabethtrykin/8sleep-mcp
- Python library `pyEight`: https://github.com/mezz64/pyEight
- Home Assistant integrations: https://github.com/lukas-clarke/eight_sleep and https://github.com/grantnedwards/eight-sleep
- Homebridge plugin: https://github.com/nfarina/homebridge-eightsleep
- Background on the unofficial API and feature removals: https://www.reddit.com/r/EightSleep/comments/15ybfrv/eight_sleep_removed_smart_home_capabilities/
