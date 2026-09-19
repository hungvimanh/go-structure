# Security Review Checklist

> OWASP/AI-LLM checklist and release-gate criteria `hero-security` walks on every pass. Fixed framework reference content, cross-referenced from [SECURITY_STANDARDS.md](./SECURITY_STANDARDS.md).

## OWASP Web Top 10 coverage

Check for:
- Broken access control: missing ownership checks, tenant leaks, unsafe direct object references, privilege escalation.
- Cryptographic failures: secrets in code/logs, weak token handling, sensitive data exposure, unsafe storage or transport.
- Injection: SQL/NoSQL/LDAP/template/command injection, unsafe shell execution, untrusted interpolation.
- Insecure design: missing abuse-case handling, unsafe trust boundaries, rate-limit gaps, missing approval gates.
- Security misconfiguration: permissive CORS, debug modes, unsafe headers, default credentials, broad permissions.
- Vulnerable/outdated components: risky dependency/config changes, unreviewed generated code or vendored code.
- Identification/authentication failures: weak session lifecycle, token refresh/logout mistakes, replay risks.
- Software/data integrity failures: unsafe deserialization, unchecked downloads, supply-chain assumptions, unsigned updates.
- Logging/monitoring failures: missing audit trails for sensitive actions, leaked secrets/PII in logs.
- SSRF and unsafe outbound access: user-controlled URLs, metadata service access, webhook fetches, redirects.

## AI / LLM feature coverage

When AI or assistant behavior is touched, also check:
- Prompt injection and tool-instruction boundary failures.
- Unauthorized tool use or data access through model-controlled paths.
- Sensitive prompt/context leakage.
- Retrieval/data poisoning assumptions.
- Unsafe generated code, command suggestions, or policy bypasses.
- Missing human approval for irreversible, external, privileged, or high-risk actions.

## Common implementation hazards

Check input validation, path traversal, upload handling, redirects, CORS/CSRF, XSS, dependency scripts, environment variable handling, logs, error messages, default-deny behavior, secret handling, API response minimization, rate limits, database grants, secure headers, mobile secure storage, and tool-calling boundaries.

## Release Gate

Block release when any of the following is true:
- Missing server-side authorization checks for sensitive actions.
- Missing input validation for external inputs.
- Secrets, credentials, tokens, private keys, or regulated data exposed in code, logs, prompts, repository, artifacts, URLs, or client storage.
- SQL/NoSQL/command/template injection path exists due to dynamic concatenation or unsafe interpolation.
- No MFA for privileged/admin accounts.
- No rate limiting or anti-bruteforce control on auth-sensitive endpoints.
- Critical dependency, runtime, image, or infrastructure vulnerability remains unresolved without documented acceptance.
- Tenant isolation is unverified for multi-tenant sensitive data.
- AI/tool-calling feature can perform irreversible, external, privileged, or high-risk action without confirmation and least privilege.
- File upload accepts executable, oversized, path-traversal, or parser-confusion payloads without effective controls.
