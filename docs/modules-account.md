# Account / Profile Module (Go)

## Purpose

This module backs the richer profile/settings surfaces used by the web dashboard and the Pistoni Swift client.

## Implemented routes

- `GET /account`
- `GET /account/overview`
- `GET/PATCH /account/settings`
- `GET/PATCH /account/safety`
- `GET/PATCH /account/privacy`
- `GET/PATCH /account/security`
- `GET/POST/PATCH/DELETE /account/trusted-contacts`
- `GET/POST/PATCH/DELETE /account/family-members`
- `GET/POST/PATCH/DELETE /account/trusted-devices`
- `GET/POST/PATCH/DELETE /account/documents`
- `GET/POST/PATCH/DELETE /account/calendars`
- `DELETE /account`

## Notes

- Route group is JWT-protected and uses the authenticated `userID` from middleware.
- Read-heavy endpoints are cached in Redis when configured.
- Mutations invalidate both account cache keys and the cached auth principal.
- Current implementation uses the in-memory repository adapter to preserve deterministic tests during migration.
