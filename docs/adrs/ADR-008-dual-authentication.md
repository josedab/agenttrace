# ADR-008: API Key + JWT Dual Authentication

## Status

Accepted

## Context

AgentTrace has two distinct authentication scenarios:

1. **Programmatic access (SDKs, CI/CD, integrations)**
   - Long-lived credentials
   - Scoped to specific projects
   - Non-interactive (no user present)
   - Needs to work in scripts, pipelines, agents

2. **Interactive access (web dashboard)**
   - User sessions with expiration
   - Multi-organization access
   - Browser-based
   - SSO integration requirements

A single authentication mechanism cannot optimally serve both scenarios. JWTs are awkward for scripts (require refresh flows), while API keys are insecure for interactive sessions (no expiration, no revocation visibility).

### Alternatives Considered

1. **API keys only**
   - Pros: Simple, universal
   - Cons: No session management, awkward for web UX, no SSO

2. **JWT only**
   - Pros: Stateless, standard
   - Cons: Token refresh complexity for scripts, no project scoping

3. **OAuth2 for everything**
   - Pros: Industry standard, supports all flows
   - Cons: Complex for simple SDK usage, overkill for agent scripts

4. **API Key + JWT** (chosen)
   - Pros: Each optimized for its use case
   - Cons: Two systems to maintain

## Decision

We implement **dual authentication** with:

1. **API Keys** for programmatic access
2. **JWT tokens** for web sessions

### API Key Design

```
Format: at_<environment>_<random>
Example: at_prod_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6

Components:
- Prefix: "at_" (AgentTrace)
- Environment: "prod", "dev", "test" (optional)
- Random: 32 character base62 string
```

**API Key Properties:**
- Scoped to a single project
- Multiple keys per project allowed
- Named keys for identification (e.g., "CI Pipeline", "Local Dev")
- Created/revoked via dashboard or API
- Stored hashed in PostgreSQL (bcrypt)
- Last used timestamp tracked

### JWT Design

```json
{
  "sub": "user_uuid",
  "org": "org_uuid",
  "email": "user@example.com",
  "role": "admin",
  "iat": 1699999999,
  "exp": 1700003599,
  "iss": "agenttrace"
}
```

**JWT Properties:**
- Short-lived access tokens (15 minutes)
- Long-lived refresh tokens (7 days)
- Issued after OAuth2/SSO login
- Contains user and organization context
- Signed with RS256 (asymmetric)

### Authentication Flow

```
┌────────────────────────────────────────────────────────────────┐
│                        Request                                  │
└────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌────────────────────────────────────────────────────────────────┐
│                  Check Authorization Header                     │
│                                                                 │
│  Bearer at_*     → API Key Authentication                      │
│  Bearer eyJ*     → JWT Authentication                          │
│  X-API-Key: *    → API Key Authentication (header)             │
│  ?api_key=*      → API Key Authentication (query, deprecated)  │
└────────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
┌─────────────────────────┐     ┌─────────────────────────┐
│   API Key Validation    │     │    JWT Validation       │
│                         │     │                         │
│ 1. Hash key             │     │ 1. Verify signature     │
│ 2. Lookup in DB         │     │ 2. Check expiration     │
│ 3. Validate project     │     │ 3. Extract claims       │
│ 4. Check if revoked     │     │ 4. Load user from DB    │
└─────────────────────────┘     └─────────────────────────┘
              │                               │
              ▼                               ▼
┌─────────────────────────┐     ┌─────────────────────────┐
│  Context: ProjectAuth   │     │  Context: UserAuth      │
│  - project_id           │     │  - user_id              │
│  - key_id               │     │  - org_id               │
│  - scopes               │     │  - email                │
└─────────────────────────┘     │  - role                 │
                                └─────────────────────────┘
```

### Middleware Implementation

```go
// auth.go
type AuthType string

const (
    AuthTypeAPIKey AuthType = "api_key"
    AuthTypeJWT    AuthType = "jwt"
)

func AuthMiddleware(keyRepo repository.APIKeyRepository, jwtService *JWTService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        auth := c.Get("Authorization")

        // Check for API key in various locations
        apiKey := extractAPIKey(c)
        if apiKey != "" {
            return authenticateAPIKey(c, keyRepo, apiKey)
        }

        // Check for JWT
        if strings.HasPrefix(auth, "Bearer eyJ") {
            return authenticateJWT(c, jwtService, strings.TrimPrefix(auth, "Bearer "))
        }

        return fiber.NewError(fiber.StatusUnauthorized, "Missing authentication")
    }
}
```

## Consequences

### Positive

- **SDK simplicity**: API keys are easy to use in scripts and SDKs
- **Web security**: JWTs provide proper session management
- **Project isolation**: API keys enforce project boundaries
- **Auditability**: Both auth types logged with request context
- **SSO ready**: JWT flow supports OAuth2/OIDC providers
- **Multiple keys**: Teams can have separate keys per environment/tool

### Negative

- **Two systems**: Must maintain both authentication paths
- **Key management**: Users must understand when to use which
- **Storage**: API keys require secure hashing and storage
- **Token refresh**: JWT clients need refresh logic

### Neutral

- API keys visible in logs must be masked
- Both auth types produce consistent context for handlers
- Rate limiting applies per key/user

## Security Considerations

### API Key Security
- Keys hashed with bcrypt (cost 12)
- Original key shown only once at creation
- Automatic key rotation reminders (90 days)
- Revocation takes effect immediately
- Keys logged as `at_****last4` for debugging

### JWT Security
- RS256 signing (asymmetric keys)
- Short access token lifetime (15 min)
- Refresh tokens stored in HttpOnly cookies
- Token revocation via blocklist in Redis
- CSRF protection for web flows

## API Examples

### SDK (API Key)
```python
from agenttrace import AgentTrace

client = AgentTrace(api_key="at_prod_a1b2c3...")
client.trace(name="my-trace")
```

### Web Dashboard (JWT)
```typescript
// After OAuth login, JWT stored in HttpOnly cookie
const response = await fetch('/api/traces', {
  credentials: 'include',  // Sends JWT cookie
});
```

### REST API (API Key Header)
```bash
curl -X POST https://api.agenttrace.io/api/public/traces \
  -H "Authorization: Bearer at_prod_a1b2c3..." \
  -H "Content-Type: application/json" \
  -d '{"name": "my-trace"}'
```

## References

- [API Key Best Practices](https://cloud.google.com/docs/authentication/api-keys)
- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)
- [OAuth 2.0 for Native Apps](https://datatracker.ietf.org/doc/html/rfc8252)
