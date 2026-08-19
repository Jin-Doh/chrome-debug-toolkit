# Security Policy

## Reporting a vulnerability

Do not open a public issue for a security vulnerability. Use the repository's
GitHub Security tab to create a private security advisory, or contact the
maintainer through the private contact option shown there.

Include:

- affected version or commit
- reproducible steps or a minimal proof of concept
- impact and likely attack prerequisites
- suggested mitigation, if known

Please allow time for validation and a coordinated fix before public
disclosure. Do not include real credentials, cookies, private URLs, or raw
NetLog captures in a report.

## Supported versions

| Version | Security fixes |
| --- | --- |
| `main` | Yes |
| older commits | No |

## Sensitive diagnostic data

Chrome NetLog, Chrome stderr, and the persistent managed profile may contain
URLs, headers, cookies, authentication-related metadata, network endpoints, or
other private browsing state. Treat every capture as sensitive, restrict file
permissions, and review or redact it before sharing.
