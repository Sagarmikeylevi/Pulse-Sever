# Task System

Pulse's core productivity tracking feature.

## Overview

The task system tracks two types of tasks:

1. **Recurring tasks** (habits) — gym, DSA, reading. Defined once, tracked daily with a rating system.
2. **Daily check-ins** — one-off tasks for today only. Created and completed within the same day.

This is not a todo app. It's a productivity tracker designed to feed into AI coaching later — every rating and note becomes training data for personalized insights.

## Design Decisions

### Two Task Types, One Table

Both types live in a single table with a `type` discriminator. The query patterns are similar enough that splitting tables would add complexity without benefit. Nullable fields are minimal — `recurrence_days` is null for check-ins, `target_date` is null for recurring.

### Rating System (Recurring Tasks Only)

Recurring tasks are rated, not checked off:

| Rating | Meaning |
|--------|---------|
| `rest` | Intentional skip — excluded from all calculations |
| `rough` | 0-25% effort range |
| `okay` | 25-50% effort range |
| `good` | 50-75% effort range |
| `crushed_it` | 75-100% effort range |

Daily check-ins use `done` only — simple completion tracking.

**Why ratings over checkboxes**: A checkbox loses context. "Did you go to the gym?" doesn't capture whether it was a great session or you left after 10 minutes. Ratings give the future AI coach meaningful signal.

### Score Column — Internal Only, Never Exposed

Each rating maps to a numeric score internally. This column is never shown to the user — it exists solely for future AI trend analysis.

**Why hidden**: If a user logs 3 consecutive "Crushed it" and sees 90%, that feels wrong — they gave the highest rating every time. User-facing analytics will use rating frequency distributions (e.g., "Crushed it 60% of the time"), not score averages.

### Recurrence: JSON Array + Daily Flag

Days are stored as a JSON array of integers (0=Sunday through 6=Saturday), matching Go's `time.Weekday()`. An `is_daily` convenience flag auto-sets when all 7 days are selected. This avoids a separate join table for a simple one-to-many relationship.

### Target Date for Check-ins

Check-ins get a server-stamped `target_date` in the user's timezone. Without this, a check-in created at 11:30 PM IST would show as the next day in UTC. The date column avoids timezone math at query time and gives a clean filter for "what did I create on September 5th?"

### One Log Per Task Per Day

Enforced at the database level with a unique constraint. Users can update a log (change rating or notes) but cannot create duplicate entries for the same day.

### No Soft Delete on Task Logs

Logs represent historical fact. If a task is deleted, cascading handles cleanup. Users edit logs (change rating/notes), they don't delete them.

### User Scoping on Every Query

Every database query includes the authenticated user's ID in the WHERE clause. Even if someone guesses another user's task UUID, the query returns nothing. Defense in depth — not relying solely on application-level checks.

## How It Works

### Creating Tasks

- **Recurring**: User provides a title and which days (or marks as daily). Recurrence days are validated (0-6 range, no duplicates). If all 7 days are provided, `is_daily` is set automatically.
- **Check-in**: User provides a title. Server stamps today's date in the user's timezone. No scheduling, no future dates.

### Logging Tasks

User selects a rating for a recurring task or marks a check-in as done. The server stamps `log_date` as today in the user's timezone and computes the internal score. If already logged today, the user gets a clear error — they should update the existing log instead.

Rating validation is type-aware: recurring tasks reject `done`, check-ins reject everything except `done`.

### Today View

The "today" endpoint combines:
- Recurring tasks where today's day-of-week (in user's timezone) matches the recurrence schedule
- Check-ins where `target_date` equals today
- Each task's log for today (if it exists)

Recurring tasks and check-ins are returned as separate arrays since the frontend renders them in different UI sections.

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| User changes timezone | Past logs keep their dates. Only future date calculations use the new timezone. |
| Check-in created near midnight | Server uses user's timezone, not UTC. 11:30 PM IST = today in IST. |
| Double-logging same task | Unique constraint blocks it. Service returns friendly error. |
| Using `done` on recurring task | Rejected — recurring tasks use the rating scale. |
| Using `good` on check-in | Rejected — check-ins only accept `done`. |
| Task deleted | Soft delete via GORM. Task logs cascade-deleted at DB level. |
| All 7 recurrence days selected | `is_daily` auto-set to true. |
| Recurrence day out of range | Rejected — must be 0-6. |
| Duplicate recurrence days | Rejected — e.g., `[1, 3, 3]` is invalid. |

## Future Work

- **Archiving**: `archived_at` column to hide tasks from daily view while preserving history for analytics.
- **Auto-failed**: Mark unlogged days as "failed" on expected recurring days. Implementation approach (cron job vs on-the-fly calculation) to be decided after CRUD is stable.
- **Analytics**: Rating distributions, completion rates, trend lines — all using rating categories, not scores.
- **AI Coach**: Chat-based analysis using rating history + notes for personalized coaching.
