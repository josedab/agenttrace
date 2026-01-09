---
sidebar_position: 1
---

# Single Sign-On (SSO)

AgentTrace Enterprise supports Single Sign-On (SSO) integration with popular identity providers, enabling seamless authentication for your organization.

## Supported Providers

| Provider | Protocol | Status |
|----------|----------|--------|
| Okta | OIDC/SAML | ✅ Supported |
| Azure AD | OIDC/SAML | ✅ Supported |
| Google Workspace | OIDC | ✅ Supported |
| Auth0 | OIDC | ✅ Supported |
| OneLogin | SAML | ✅ Supported |
| Custom SAML 2.0 | SAML | ✅ Supported |
| Custom OIDC | OIDC | ✅ Supported |

## Quick Setup

### 1. Navigate to SSO Settings

1. Go to **Settings > Organization > SSO**
2. Click **Configure SSO**
3. Select your identity provider

### 2. Configure Your IdP

Follow the provider-specific guides below to configure your identity provider.

### 3. Test and Enable

1. Click **Test Connection** to verify the configuration
2. Optionally enable **Enforce SSO** to require SSO for all users
3. Click **Enable SSO**

## Provider Guides

### Okta

#### Step 1: Create an Application in Okta

1. Log in to your Okta Admin Console
2. Go to **Applications > Applications > Create App Integration**
3. Select **OIDC - OpenID Connect**
4. Select **Web Application**
5. Click **Next**

#### Step 2: Configure the Application

| Setting | Value |
|---------|-------|
| App integration name | AgentTrace |
| Sign-in redirect URIs | `https://app.agenttrace.io/api/auth/sso/callback` |
| Sign-out redirect URIs | `https://app.agenttrace.io` |
| Assignments | Assign to appropriate users/groups |

#### Step 3: Get Configuration Values

From the Okta application, copy:
- **Client ID**
- **Client Secret**
- **Okta Domain** (e.g., `your-company.okta.com`)

#### Step 4: Configure in AgentTrace

```json
{
  "provider": "okta",
  "oidcClientId": "your-client-id",
  "oidcClientSecret": "your-client-secret",
  "oidcIssuerUrl": "https://your-company.okta.com",
  "oidcScopes": ["openid", "profile", "email", "groups"]
}
```

### Azure AD

#### Step 1: Register an Application

1. Go to **Azure Portal > Azure Active Directory > App registrations**
2. Click **New registration**
3. Configure:
   - Name: `AgentTrace`
   - Supported account types: Choose based on your needs
   - Redirect URI: `https://app.agenttrace.io/api/auth/sso/callback`

#### Step 2: Configure Authentication

1. Go to **Authentication**
2. Add platform: **Web**
3. Enable **ID tokens** under Implicit grant

#### Step 3: Create Client Secret

1. Go to **Certificates & secrets**
2. Click **New client secret**
3. Copy the secret value immediately

#### Step 4: Configure in AgentTrace

```json
{
  "provider": "azure_ad",
  "oidcClientId": "your-application-id",
  "oidcClientSecret": "your-client-secret",
  "oidcIssuerUrl": "https://login.microsoftonline.com/your-tenant-id/v2.0",
  "oidcScopes": ["openid", "profile", "email"]
}
```

### Google Workspace

#### Step 1: Create OAuth Credentials

1. Go to **Google Cloud Console > APIs & Services > Credentials**
2. Click **Create Credentials > OAuth client ID**
3. Select **Web application**
4. Add authorized redirect URI: `https://app.agenttrace.io/api/auth/sso/callback`

#### Step 2: Configure in AgentTrace

```json
{
  "provider": "google",
  "oidcClientId": "your-client-id.apps.googleusercontent.com",
  "oidcClientSecret": "your-client-secret",
  "oidcIssuerUrl": "https://accounts.google.com",
  "oidcScopes": ["openid", "profile", "email"]
}
```

### Custom SAML 2.0

#### Step 1: Get AgentTrace Metadata

Download the AgentTrace SP metadata from:
```
https://app.agenttrace.io/api/auth/saml/metadata?orgId=YOUR_ORG_ID
```

Or use these values manually:

| Setting | Value |
|---------|-------|
| Entity ID | `https://app.agenttrace.io/saml/YOUR_ORG_ID` |
| ACS URL | `https://app.agenttrace.io/api/auth/sso/saml/callback` |
| SLO URL | `https://app.agenttrace.io/api/auth/sso/saml/logout` |

#### Step 2: Configure Your IdP

Configure your SAML IdP with the above values and ensure it sends these attributes:

| Attribute | Description |
|-----------|-------------|
| `email` | User's email address (required) |
| `firstName` | User's first name |
| `lastName` | User's last name |
| `groups` | User's group memberships |

#### Step 3: Configure in AgentTrace

```json
{
  "provider": "saml",
  "samlEntityId": "https://your-idp.com/entity-id",
  "samlSsoUrl": "https://your-idp.com/sso",
  "samlSloUrl": "https://your-idp.com/slo",
  "samlCertificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
  "samlSignRequests": true,
  "samlNameIdFormat": "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
}
```

## Configuration Options

### Attribute Mapping

Map IdP attributes to AgentTrace user fields:

```json
{
  "attributeMapping": {
    "email": "email",
    "firstName": "given_name",
    "lastName": "family_name",
    "displayName": "name",
    "groups": "groups",
    "department": "department"
  }
}
```

### Domain Restrictions

Limit SSO to specific email domains:

```json
{
  "allowedDomains": ["your-company.com", "subsidiary.com"]
}
```

### Auto-Provisioning

Automatically create accounts for new SSO users:

```json
{
  "autoProvisionUsers": true,
  "defaultRole": "member",
  "autoAssignProjects": ["project-id-1", "project-id-2"]
}
```

### Enforce SSO

Require all users to authenticate via SSO:

```json
{
  "enforceSSO": true
}
```

When enabled:
- Users cannot log in with email/password
- Existing sessions remain valid until expiry
- Admins can still bypass SSO via a special URL

## API Configuration

### Get SSO Configuration

```bash
GET /v1/organizations/:orgId/sso/config
```

### Configure SSO

```bash
PUT /v1/organizations/:orgId/sso/config
Content-Type: application/json

{
  "provider": "oidc",
  "enabled": true,
  "enforceSSO": false,
  "allowedDomains": ["your-company.com"],
  "oidcClientId": "...",
  "oidcClientSecret": "...",
  "oidcIssuerUrl": "https://your-idp.com",
  "oidcScopes": ["openid", "profile", "email"],
  "attributeMapping": {
    "email": "email",
    "firstName": "given_name",
    "lastName": "family_name"
  },
  "autoProvisionUsers": true,
  "defaultRole": "member"
}
```

### Enable/Disable SSO

```bash
POST /v1/organizations/:orgId/sso/enable
POST /v1/organizations/:orgId/sso/disable
```

## Troubleshooting

### Common Issues

#### "Invalid state parameter"

- The SSO flow took too long (> 10 minutes)
- User navigated away and back
- **Solution**: Start the SSO flow again

#### "Email domain not allowed"

- User's email domain isn't in `allowedDomains`
- **Solution**: Add the domain or remove domain restrictions

#### "User not found and auto-provisioning disabled"

- New user trying to log in
- `autoProvisionUsers` is false
- **Solution**: Enable auto-provisioning or create the user manually

#### "SAML signature validation failed"

- IdP certificate is incorrect or expired
- **Solution**: Update the SAML certificate

### Debug Logging

Enable SSO debug logging:

```bash
# Self-hosted
export AGENTTRACE_SSO_DEBUG=true
```

View SSO errors in the dashboard:
1. Go to **Settings > Organization > SSO**
2. Check **Last Error** section

## Security Best Practices

1. **Use HTTPS**: Always use HTTPS for redirect URIs
2. **Rotate Secrets**: Rotate client secrets periodically
3. **Audit Logs**: Monitor SSO events in audit logs
4. **MFA**: Enable MFA in your identity provider
5. **Session Timeout**: Configure appropriate session timeouts
6. **Domain Restrictions**: Use `allowedDomains` to prevent unauthorized access

## Session Management

### Session Duration

SSO sessions inherit the token lifetime from your IdP. AgentTrace also enforces:

| Setting | Default | Description |
|---------|---------|-------------|
| Access token lifetime | 1 hour | Short-lived access |
| Refresh token lifetime | 7 days | For token refresh |
| Idle timeout | 24 hours | Inactivity timeout |

### Session Termination

Sessions are terminated when:
- User explicitly logs out
- Token expires and refresh fails
- Admin revokes the session
- User is removed from the IdP

### Force Logout

Terminate all SSO sessions for a user:

```bash
POST /v1/organizations/:orgId/users/:userId/sessions/terminate
```
