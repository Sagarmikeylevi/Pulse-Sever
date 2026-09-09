# Timezone System

Per-user timezone storage and change detection for Pulse.

## Why timezone matters

Task dates, auto-failed logic, and future AI coaching all depend on knowing the user's local date. Without a stored timezone, the server can't determine when a user's "day" starts and ends.

## Design Decisions

### IANA timezone names, not UTC offsets

We store `"Asia/Kolkata"`, not `"+05:30"`. UTC offsets change with DST — `"America/New_York"` is -05:00 in winter and -04:00 in summer. IANA names handle this automatically.

### Required on signup, never auto-updated

- Timezone is captured during the signup flow. The client detects it from the device and sends it in the request body.
- The server never auto-updates timezone. When a mismatch is detected, the client prompts the user to confirm.
- This prevents false updates from VPNs, short trips, or multi-device usage.

### Server-side validation

All timezone strings are validated server-side using Go's `time.LoadLocation()`. Invalid IANA names are rejected. The `time/tzdata` package is embedded in the binary for portability in containerized environments where the OS may lack timezone data.

### Historical data integrity

When a user changes timezone, past task logs are NOT retroactively adjusted. A task logged on `2026-09-05` stays on that date regardless of future timezone changes. Only future date calculations use the new timezone.

## Timezone change detection

The client calls a check endpoint on app open, sending the device's current timezone. The server compares it against the stored value and returns whether they match, along with both values. If mismatched, the client decides when to prompt (e.g., only after persistent mismatch across multiple app opens over 24+ hours to avoid nagging during short trips).

The user must explicitly confirm to update their timezone — the server never changes it silently.

## Edge cases

| Scenario | Behavior |
|----------|----------|
| Travel (short trip) | Client-side logic: only prompt after persistent mismatch |
| VPN / spoofed timezone | User confirms or dismisses — no auto-update |
| Multiple devices | Server stores ONE timezone — the one user last confirmed |
| Date goes backward on change | Already-logged tasks keep their date. User gets extra time on the "old" day. No data loss. |
| Date goes forward on change | Incomplete tasks for the skipped period become candidates for auto-failed. Same as if user didn't open the app. |
| DST transition | Handled automatically by IANA timezone names |
| Missing timezone on signup | Rejected — timezone is required |
| Invalid timezone string | Rejected — validated via `time.LoadLocation()` |
