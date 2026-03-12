# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| latest  | ✅ Yes             |
| < latest | ❌ No (upgrade recommended) |

We recommend always running the latest version of AgentTrace. Security patches are applied to the latest release only.

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

If you discover a security vulnerability, please report it responsibly using one of these methods:

1. **GitHub Security Advisories (preferred)**: [Report a vulnerability](https://github.com/agenttrace/agenttrace/security/advisories/new)
2. **Email**: Send details to [security@agenttrace.io](mailto:security@agenttrace.io)

### What to Include

- A description of the vulnerability and its potential impact
- Step-by-step instructions to reproduce the issue
- Affected versions and components (API, web, SDK, etc.)
- Any potential mitigations you've identified

### What to Expect

| Step | Timeline |
|------|----------|
| Acknowledgment of your report | Within 48 hours |
| Initial assessment and severity rating | Within 1 week |
| Fix development and testing | Depends on severity |
| Security advisory and patch release | Coordinated with reporter |

We will keep you informed of progress throughout the process.

## Scope

The following are in scope for security reports:

- **API server** (`api/`) — authentication, authorization, injection, data leaks
- **Web application** (`web/`) — XSS, CSRF, session management
- **SDKs** (`sdk/`) — credential handling, data exposure
- **Deployment configurations** (`deploy/`) — default credentials, insecure defaults
- **Dependencies** — known vulnerabilities in third-party packages

## Security Best Practices

When self-hosting AgentTrace:

- Generate strong secrets for `JWT_SECRET`, `ENCRYPTION_KEY`, and `NEXTAUTH_SECRET` using `openssl rand -base64 32`
- Never use default passwords in production
- Enable TLS/HTTPS for all external-facing services
- Restrict network access to database ports (PostgreSQL, ClickHouse, Redis, MinIO)
- Regularly update to the latest version
- Review the [deployment guide](api/docs/DEPLOYMENT.md) for production hardening recommendations

## Acknowledgments

We appreciate the security research community's efforts in responsibly disclosing vulnerabilities. Contributors who report valid security issues will be acknowledged in our release notes (unless they prefer to remain anonymous).
