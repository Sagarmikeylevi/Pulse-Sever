# Authentication System

## Overview

Pulse uses a dual-path authentication system:

1. **Email + OTP** — required for first-time users (sign-up), always available for login
2. **Email + Password** — available after a user sets a password post-OTP verification

First-time users must verify their email via OTP before gaining access. Once verified, they can optionally set a password for future logins.

## Design Decisions

### Token Strategy: JWT Access Token + Refresh Token

- **Access token**: JWT signed with HS256, 15-minute expiry. Stateless — no DB lookup needed.
- **Refresh token**: Opaque string, hashed and stored in DB, 30-day expiry. Revocable server-side.
- **Why HS256 over RS256**: Single-service architecture. The Go server is the only issuer and verifier. The planned LLM service will be called internally, never needing to verify JWTs. HS256 is simpler with no security tradeoff here.

### Refresh Token Rotation with Family-Based Theft Detection

Each login creates a token "family" (shared UUID). On every refresh:

1. Old token is marked `used`, new one issued in the same family.
2. If a `used` token is replayed → two parties hold copies → theft detected.
3. Entire family is revoked, forcing re-authentication.

**Why**: A 30-day refresh token is a high-value target. Rotation limits a stolen token to a single use, and family tracking enables detection.

### OTP Storage: PostgreSQL (not Redis)

**Why not Redis**: Adding Redis introduces another infra dependency for data that's short-lived and low-volume. PostgreSQL handles OTP queries in <5ms at Pulse's scale. Redis will be introduced when there's a measured need (e.g., caching), not a hypothetical one.

### OTP Codes: Hashed with bcrypt

Even though OTPs are short-lived (5 min), a DB breach could expose active codes. Hashing eliminates this risk.

### Password Hashing: bcrypt

Via `golang.org/x/crypto`.

### Separate Profile Table (planned)

Auth data (`email`, `password_hash`, `is_email_verified`) lives in `users`. Profile data (`name`, `avatar`, etc.) will live in a separate `profiles` table. Auth fields are read on every request; profile fields are not. Keeps auth queries lean.

### UUID Primary Keys

UUIDs don't leak information (can't infer user count from ID) and work better if the architecture becomes distributed.

### Soft Deletes

GORM's `deleted_at` timestamp. Preserves audit trails, allows recovery, excluded from normal queries automatically.

## Auth Flows

### Sign Up (new user)

1. Client sends email → server generates OTP, hashes and stores it, sends via email (console in dev)
2. Client submits OTP + timezone → server verifies OTP (hash, expiry, attempts), creates user with timezone, returns tokens
3. Client optionally sets a password (authenticated)

### Login with OTP (returning user)

Same as signup flow, but the user already exists — server finds them instead of creating.

90-second cooldown between OTP requests per email.

### Login with Password

Client sends email + password → server verifies with bcrypt → returns tokens.

### Token Refresh

Client sends refresh token → server looks up hash → if active, rotates (mark used, issue new pair in same family). If the token was already `used`, theft is detected and the entire family is revoked.

### Forgot Password

No separate reset flow. User logs in via OTP, then calls set-password endpoint while authenticated.

## OTP Rules

- 6-digit code, cryptographically generated
- 5-minute expiry
- Max 3 verification attempts per OTP
- 90-second cooldown between sends
- Hashed with bcrypt before storage

## Email Delivery

- **Development**: OTP codes logged to console
- **Production**: Sent via AWS SES
- Controlled by `APP_ENV` environment variable
