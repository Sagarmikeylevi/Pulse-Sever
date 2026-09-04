# Authentication System

## Overview

Pulse uses a dual-path authentication system:

1. **Email + OTP** — required for first-time users (sign-up), always available for login
2. **Email + Password** — available after a user sets a password post-OTP verification

First-time users must verify their email via OTP before gaining access. Once verified, they can optionally set a password for future logins.

## Design Decisions

### Token Strategy: JWT Access Token + Refresh Token

- **Access token**: JWT signed with HS256, 15-minute expiry. Stateless — validated by checking the signature, no DB lookup.
- **Refresh token**: Opaque string, hashed and stored in the DB, 30-day expiry. Can be revoked server-side.
- **Why HS256 over RS256**: Pulse is a single-service architecture. The Go server is the only service that issues and verifies tokens. The LLM service (planned) will be called internally by the Go server, so it never needs to verify JWTs independently. HS256 is simpler with no security tradeoff in this architecture.

### Refresh Token Rotation with Family-Based Theft Detection

Each login creates a token "family" (shared UUID). On every refresh:

1. Old refresh token is marked `used`, new one is issued in the same family.
2. If a `used` token is replayed, it means two parties hold copies of the same token — theft detected.
3. The entire family is revoked, forcing re-authentication.

**Why**: A long-lived refresh token (30 days) is a high-value target. Rotation limits the window of a stolen token to a single use, and family tracking enables detection.

### OTP Storage: PostgreSQL (not Redis)

OTPs are stored in a PostgreSQL table with an `expires_at` column.

**Why not Redis**: Redis is ideal for short-lived key-value data, but adding it introduces another infrastructure dependency (docker-compose service, connection management, failure modes). At Pulse's scale, PostgreSQL handles OTP queries in <5ms. Redis will be introduced when there's a measured need (e.g., API response caching), not a hypothetical one.

### OTP Codes: Hashed

OTP codes are hashed before storage using bcrypt.

**Why**: Even though OTPs are short-lived (5 min), a database breach could expose active codes. Hashing eliminates this risk at negligible performance cost.

### Password Hashing: bcrypt

Passwords are hashed using bcrypt via `golang.org/x/crypto` (already an indirect dependency).

### Separate Profile Table

User auth data (`email`, `password_hash`, `is_email_verified`) lives in the `users` table. Profile data (`name`, `avatar`, etc.) will live in a separate `profiles` table.

**Why**: Auth fields are read on every authenticated request. Profile fields are read only when viewing a profile. Keeping the user table lean keeps auth queries fast, and the two concerns evolve independently.

### Primary Keys: UUID

All tables use UUID primary keys.

**Why**: UUIDs don't leak information (an attacker can't infer user count from an ID) and work better in distributed systems if the architecture evolves.

### Soft Deletes

Core tables use GORM's soft delete (`deleted_at` timestamp) instead of hard deletes.

**Why**: Preserves data for audit trails and allows recovery. Deleted rows are excluded from normal queries automatically by GORM.

## Auth Flows

### Sign Up (new user)

```
1. Client sends email         → POST /auth/otp/send
2. Server generates OTP       → hashes and stores in DB, logs to console (dev) or sends via SES (prod)
3. Client submits OTP          → POST /auth/otp/verify
4. Server verifies OTP         → checks hash, expiry (5 min), attempts (max 3)
5. OTP valid                   → creates user (is_email_verified: true), returns access + refresh tokens
6. Client sets password (opt)  → POST /auth/password/set (authenticated)
7. Server hashes and stores    → user now has both login paths available
```

### Login with OTP (returning user)

```
1. Client sends email         → POST /auth/otp/send
2. Server checks cooldown     → rejects if <90 sec since last OTP for this email
3. Server generates OTP       → same as sign-up flow
4. Client submits OTP          → POST /auth/otp/verify
5. Server verifies OTP         → user exists, returns access + refresh tokens
```

### Login with Password

```
1. Client sends email + password → POST /auth/login
2. Server finds user by email    → checks password hash with bcrypt
3. Password valid                → returns access + refresh tokens
```

### Token Refresh

```
1. Client sends refresh token    → POST /auth/token/refresh
2. Server hashes and looks up    → finds token in DB
3. Token is "active"             → marks it "used", issues new access + refresh token (same family)
4. Token is "used"               → THEFT DETECTED — revokes entire family
5. Token is "revoked"/not found  → rejected
```

### Forgot Password

No separate reset flow. The user logs in via OTP and then uses the set password endpoint:

```
1. Client sends email           → POST /auth/otp/send
2. Client verifies OTP          → POST /auth/otp/verify → returns tokens
3. Client sets new password     → POST /auth/password/set (authenticated)
```

## API Endpoints

| Method | Path | Auth Required | Description |
|--------|------|---------------|-------------|
| POST | /auth/otp/send | No | Send OTP to email |
| POST | /auth/otp/verify | No | Verify OTP, returns tokens |
| POST | /auth/login | No | Login with email + password |
| POST | /auth/password/set | Yes | Set or change password |
| POST | /auth/token/refresh | No | Exchange refresh token for new token pair |

## Database Schema

### users

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | UUID | PK | Generated by server |
| email | VARCHAR(255) | UNIQUE, NOT NULL | |
| password_hash | VARCHAR(255) | NULLABLE | Null until user sets a password |
| is_email_verified | BOOLEAN | NOT NULL, DEFAULT false | Set true after OTP verification |
| created_at | TIMESTAMPTZ | NOT NULL | |
| updated_at | TIMESTAMPTZ | NOT NULL | |
| deleted_at | TIMESTAMPTZ | NULLABLE | Soft delete |

### otps

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | UUID | PK | |
| email | VARCHAR(255) | NOT NULL | Not a FK — user may not exist yet during sign-up |
| code_hash | VARCHAR(255) | NOT NULL | bcrypt hash of OTP code |
| attempts | INTEGER | NOT NULL, DEFAULT 0 | Max 3 before invalidation |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'active' | active, used, expired |
| expires_at | TIMESTAMPTZ | NOT NULL | created_at + 5 minutes |
| created_at | TIMESTAMPTZ | NOT NULL | Used for 90-sec cooldown check |

### refresh_tokens

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | UUID | PK | |
| user_id | UUID | FK → users.id, NOT NULL | |
| token_hash | VARCHAR(255) | NOT NULL | Hashed refresh token |
| family_id | UUID | NOT NULL | Groups rotation chain for theft detection |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'active' | active, used, revoked |
| expires_at | TIMESTAMPTZ | NOT NULL | created_at + 30 days |
| created_at | TIMESTAMPTZ | NOT NULL | |

### Indexes

- `users.email` — unique index (implicit from UNIQUE constraint)
- `otps.email` — for lookup during verification and cooldown check
- `refresh_tokens.token_hash` — for lookup during token refresh
- `refresh_tokens.family_id` — for revoking entire family on theft detection
- `refresh_tokens.user_id` — for revoking all tokens on password reset

## Key Files

```
internal/
├── entity/
│   ├── user.go                        # User struct
│   ├── otp.go                         # OTP struct
│   └── refresh_token.go               # RefreshToken struct
├── repository/
│   ├── user_repository.go             # Interface + impl: FindByEmail, Create, UpdatePassword
│   ├── otp_repository.go              # Interface + impl: Create, FindActiveByEmail, IncrementAttempts
│   └── refresh_token_repository.go    # Interface + impl: Create, FindByHash, RevokeFamily
├── service/
│   ├── auth_service.go                # Interface + impl: SignUp, Login, VerifyOTP, RefreshToken
│   ├── email_service.go               # Interface + impl: SendOTP (console in dev, SES in prod)
│   └── token_service.go               # Interface + impl: GenerateJWT, ValidateJWT, GenerateRefreshToken
├── controller/
│   └── auth_controller.go             # HTTP handlers for all auth endpoints
├── middleware/
│   └── auth_middleware.go             # JWT extraction and validation from Authorization header
├── routes/
│   ├── auth_routes.go                 # Maps /auth/* endpoints to controller methods
│   └── router.go                      # Main Gin router setup
└── registry/
    └── registry.go                    # Dependency wiring

migration/
├── 000001_create_users_table.up.sql
├── 000001_create_users_table.down.sql
├── 000002_create_otps_table.up.sql
├── 000002_create_otps_table.down.sql
├── 000003_create_refresh_tokens_table.up.sql
└── 000003_create_refresh_tokens_table.down.sql
```

## Email Delivery

- **Development**: OTP codes are logged to the console (`[DEV] OTP for user@email.com: 482937`)
- **Production**: OTP codes are sent via AWS SES
- Controlled by the `APP_ENV` environment variable
