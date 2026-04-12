# Pistoni NestJS -> Go Endpoint Parity Matrix

## Scope covered in this migration pass

- ✅ Auth module
- ✅ Cars (+ car type access)
- ✅ Car-owner related flows (users + cars by owner)
- ✅ Bodyguards flows (within Providers module)

> Base path in Go is **same as NestJS controllers** (no `/api/v1` prefix) to preserve frontend compatibility.

## Auth (`src/auth/auth.controller.ts`)

| Endpoint | NestJS | Go | Status |
|---|---|---|---|
| `POST /auth/login` | login | implemented | ✅ |
| `POST /auth/google` | google oauth | implemented (token verification TODO adapter) | ⚠️ |
| `POST /auth/apple` | apple oauth | implemented (signature verification TODO adapter) | ⚠️ |
| `POST /auth/refresh` | refresh token | implemented | ✅ |
| `POST /auth/logout` | jwt protected | implemented | ✅ |
| `POST /auth/check-email` | step 1 registration | implemented | ✅ |
| `POST /auth/validate-email` | step 2 registration | implemented | ✅ |
| `POST /auth/complete-registration` | step 3 registration | implemented | ✅ |
| `POST /auth/resend-code` | resend code | implemented | ✅ |
| `POST /auth/register` | legacy registration | implemented | ✅ |
| `POST /auth/verify-code` | legacy verify code | implemented | ✅ |
| `GET /auth/verify-email/:token` | legacy token verify | implemented | ✅ |
| `GET /auth/profile` | jwt protected | implemented | ✅ |
| `PATCH /auth/profile` | jwt protected | implemented | ✅ |
| `POST /auth/change-password` | jwt protected | implemented | ✅ |

## Cars (`src/cars/cars.controller.ts`)

| Endpoint | NestJS | Go | Status |
|---|---|---|---|
| `POST /cars` | create | implemented | ✅ |
| `POST /cars/bulk` | bulk create | implemented | ✅ |
| `GET /cars` | list with `skip/take/status` | implemented | ✅ |
| `GET /cars/status/:status` | filter by status | implemented | ✅ |
| `GET /cars/owner/:id` | owner cars | implemented | ✅ |
| `GET /cars/type/:vehicleType` | filter vehicle type | implemented | ✅ |
| `GET /cars/:id` | single car | implemented | ✅ |
| `PATCH /cars/:id` | update car | implemented | ✅ |
| `DELETE /cars/:id` | delete car | implemented | ✅ |

### Added convenience endpoint
- `GET /cars/types` (returns `SUV|COASTER|LONG_VEHICLE`)

## Users / car-owner support (`src/users/users.controller.ts`)

| Endpoint | NestJS | Go | Status |
|---|---|---|---|
| `POST /users` | create user | implemented | ✅ |
| `GET /users` | list with `skip/take/role` | implemented | ✅ |
| `GET /users/role/:role` | role filter | implemented | ✅ |
| `GET /users/email/:email` | find by email | implemented | ✅ |
| `GET /users/:id` | find one | implemented | ✅ |
| `PATCH /users/:id` | update | implemented | ✅ |
| `DELETE /users/:id` | remove | implemented | ✅ |

## Providers bodyguards (`src/providers/providers.controller.ts`)

| Endpoint | NestJS | Go | Status |
|---|---|---|---|
| `POST /providers/body-guards/register` | jwt protected | implemented | ✅ |
| `GET /providers/body-guards` | query list | implemented | ✅ |
| `GET /providers/body-guards/me/profile` | jwt protected | implemented | ✅ |
| `PATCH /providers/body-guards/me/availability` | jwt protected | implemented | ✅ |
| `PATCH /providers/body-guards/:id/verification` | admin-only | implemented | ✅ |
| `GET /providers/body-guards/:id` | details | implemented | ✅ |
| `POST /providers/body-guards/join` | queue admission | implemented | ✅ |
| `GET /providers/body-guards/queue` | admin queue list | implemented | ✅ |
| `POST /providers/body-guards/queue/:id/approve` | admin approve | implemented | ✅ |
| `POST /providers/body-guards/queue/:id/reject` | admin reject | implemented | ✅ |
| `GET /providers/body-guards/me/status` | provider realtime status | implemented | ✅ |
| `PATCH /providers/body-guards/me/status` | provider realtime status | implemented | ✅ |
| `GET /providers/body-guards/:id/status` | public status read | implemented | ✅ |

## Providers drivers (phase 2 additions)

| Endpoint | NestJS style parity | Go | Status |
|---|---|---|---|
| `GET /providers/drivers/me/status` | profile-scoped status | implemented | ✅ |
| `PATCH /providers/drivers/me/status` | profile-scoped status update | implemented | ✅ |
| `GET /providers/drivers/:id/status` | status retrieval | implemented | ✅ |
| `GET /providers/drivers/status/updates` | polling updates by version/timestamp | implemented | ✅ |
| `GET /providers/drivers/:id/reviews` | list reviews + aggregate | implemented | ✅ |
| `POST /providers/drivers/:id/reviews` | create review | implemented | ✅ |
| `GET /providers/drivers/:id/counters` | assignment/route counters (admin) | implemented | ✅ |
| `PATCH /providers/drivers/:id/counters` | assignment/route counter deltas (admin) | implemented | ✅ |

## Account / profile additions

| Endpoint | Purpose | Go | Status |
|---|---|---|---|
| `GET /account` | aggregated account/profile overview | implemented | ✅ |
| `GET /account/overview` | explicit overview alias | implemented | ✅ |
| `GET/PATCH /account/settings` | language, timezone, notification prefs, text size, accessibility | implemented | ✅ |
| `GET/PATCH /account/safety` | home city, emergency protocol, live location, SOS | implemented | ✅ |
| `GET/PATCH /account/privacy` | privacy and sharing controls | implemented | ✅ |
| `GET/PATCH /account/security` | 2FA flags, session alerts, login notifications | implemented | ✅ |
| `GET/POST/PATCH/DELETE /account/trusted-contacts` | emergency/trusted contacts CRUD | implemented | ✅ |
| `GET/POST/PATCH/DELETE /account/family-members` | family access/contact CRUD | implemented | ✅ |
| `GET/POST/PATCH/DELETE /account/trusted-devices` | trusted device CRUD | implemented | ✅ |
| `GET/POST/PATCH/DELETE /account/documents` | personal document metadata CRUD | implemented | ✅ |
| `GET/POST/PATCH/DELETE /account/calendars` | calendar sync metadata CRUD | implemented | ✅ |
| `DELETE /account` | authenticated account deletion | implemented | ✅ |

## Security middleware (global)

- IP + route-aware rate limiter is applied globally.
- Sensitive mutating providers routes are protected by admission queue middleware.

## Prisma Go integration status

- `core/prisma/schema.prisma` maintained for `prisma-client-go` generation.
- Runtime in this pass uses repository interfaces + in-memory adapters for deterministic tests.
- Redis now backs profile/account JSON caching and cache invalidation when configured.
- TODO noted in auth OAuth verification path for production-grade provider token verification and Prisma-backed repository adapters.
