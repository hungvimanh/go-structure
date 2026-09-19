# Security Control Baseline

> The 14-domain control checklist `hero-security` grades every review against. Fixed framework reference content — not project-specific like [SECURITY_STANDARDS.md](./SECURITY_STANDARDS.md), which holds this project's own filled-in policy.

Use this baseline to classify controls as **Required**, **Missing**, **Verified**, **Not applicable**, or **Not reviewed**.

## 1) Input Handling and Data Processing

Required:
- Validate all client input at the server boundary: type, length, range, format, regex, and allow-list where possible.
- Never trust client data, including hidden fields, UI state, and client-side validation results.
- Encode or escape data before use in HTML, JS, CSS, SQL, LDAP, template, shell, and log contexts.
- Prefer allow-list validation over deny-list validation.

Verification:
- Negative tests cover malformed, oversized, unexpected-type, and boundary input.
- No dynamic query, template, shell, or expression injection path accepts raw input.

## 2) Authentication and Accounts

Required:
- Use established standards such as SSO, OAuth2, or OIDC where possible.
- Store passwords only with salted bcrypt, Argon2, or scrypt; never plaintext or reversible encryption.
- Do not hardcode credentials or secrets in source code, config, generated examples, logs, or repository history.
- Require MFA for privileged/admin accounts.
- Enforce first-login password change for internal or privileged accounts when password login exists.
- Privileged/internal credential rotation should be at most 90 days, and privileged/admin accounts cannot reuse the last 5 passwords.
- Privileged/service account credentials must be strong random values protected by a secret manager.
- Lockout policy: internal/admin accounts lock after 6 failed attempts in 15 minutes; customer accounts use progressive delay and captcha or equivalent abuse protection.
- Password policy: internal/admin minimum 12 characters, privileged minimum 14 with upper/lower/digit/special; customer minimum 8 and recommended 10+ with strength checks and 2FA capability.

Verification:
- Tests cover lockout, MFA path, password reuse, rotation, and generic auth failure behavior.

## 3) Session Management

Required:
- Session IDs must be cryptographically random and high entropy.
- Cookies must set `Secure`, `HttpOnly`, and `SameSite=Strict` or `SameSite=Lax`.
- Session invalidation must occur server-side on logout, privilege changes, and credential reset.
- Centralize auth/session controls in shared middleware or a shared library.
- Enforce inactivity timeout according to risk level; 15 minutes or less is recommended for sensitive/admin flows.
- Never send session IDs, access tokens, or refresh tokens through URLs or logs.

Verification:
- Evidence confirms server-side revocation, timeout behavior, cookie flags, and no token leakage in URLs/logs.

## 4) Access Control

Required:
- Enforce authorization server-side only; UI controls are never security controls.
- Deny by default and grant only required permissions.
- Apply least privilege and role separation across user, admin, privileged, and service accounts.
- No shared privileged accounts.
- Never create a default super admin during bootstrap.
- Privileged account lifecycle must be approved, time-bounded, logged, and promptly revoked.
- Multi-tenant systems must enforce tenant isolation on every object read, write, export, webhook, background job, and admin path.

Verification:
- Tests cover horizontal and vertical privilege escalation, direct object references, tenant leaks, and sensitive action authorization.

## 5) Cryptography and Data Protection

Required:
- Use standard algorithms and maintained libraries only, such as AES-256, RSA-2048+, or ECC where appropriate.
- Never create custom cryptographic algorithms or ad-hoc signing/encryption formats.
- Manage keys in Vault or a Secret Manager with rotation policy and scoped access.
- Protect data stored on web/mobile clients: minimize retention, delete when unnecessary, encrypt or mask sensitive data.
- Store mobile credentials/secrets only in OS secure stores such as Keychain or Keystore.
- Never store auth tokens in `localStorage` or `sessionStorage`; use HttpOnly cookies or secure ephemeral memory.
- Protect sensitive data at rest and in transit according to classification.

Verification:
- Evidence exists for key source, key access scope, rotation, TLS, encrypted backups, and sensitive data minimization.

## 6) Error Handling and Logging

Required:
- Return generic authentication errors; do not reveal user existence or exact failure reason.
- Log critical security events: login, logout, failed login, MFA changes, permission changes, credential resets, exports, and sensitive create/update/delete actions.
- Use structured logs such as JSON or syslog-compatible records.
- Do not expose stack traces, system paths, versions, or internal error details to end users.
- Never log passwords, secrets, tokens, session IDs, prompt contents containing sensitive data, or sensitive PII.
- Audit important business actions with actor, timestamp, source IP, action, target object, and result.
- Applications must not be able to modify or delete their own immutable security logs.

Verification:
- Sampling confirms logs are structured, useful for audit, and free of sensitive fields.

## 7) Database Security

Required:
- Use prepared statements or safe ORM patterns; no string concatenation SQL or dynamic query construction from raw input.
- Database accounts must follow least privilege; applications must not use root, `sa`, owner, or broad admin accounts.
- Enable audit logging for sensitive tables and sensitive actions such as CREATE, DROP, ALTER, UPDATE, DELETE, and INSERT, then forward to SIEM or equivalent monitoring.
- Encrypted backups are mandatory for sensitive systems.
- Sensitive columns or tables must be encrypted or tokenized where classification requires it.
- AI internal apps must have scoped DB/knowledge access, dedicated service accounts, and access logging.
- Each service can access only its own database; avoid cross-service shared DB accounts.

Verification:
- Review grants, SQL patterns, audit pipeline, backup encryption, and service account scope.

## 8) File Handling and Upload

Required when upload or file processing exists:
- Validate MIME type, extension, and magic number.
- Store uploaded files outside webroot or outside direct executable/static serving paths.
- Integrate malware scanning or sandboxing where applicable.
- Enforce file size, count, and processing time limits.
- Block executable uploads such as `.php`, `.jsp`, `.exe`, scripts, and dangerous archive formats unless explicitly justified and sandboxed.
- Generate non-predictable server-side filenames; never trust user-provided names or paths.
- Prevent path traversal, archive slip, symlink traversal, parser confusion, and polyglot bypasses.

Verification:
- Tests include renamed executable, polyglot file, oversized file, archive traversal, MIME mismatch, and path traversal attempts.

## 9) Communication and API Security

Required:
- Expose only necessary APIs and maintain clear documentation.
- Rotate API keys and tokens periodically.
- Apply rate limiting and anti-bruteforce controls for APIs, login endpoints, password reset, MFA, exports, and expensive operations.
- Never send access tokens or secrets in URLs.
- Do not expose version banners, software details, stack traces, or debug data in responses.
- Use explicit API versioning; do not silently break previous behavior.
- Return minimum required fields only; avoid over-fetching full objects.
- Validate webhook signatures and replay protections.

Verification:
- Contract tests confirm response field minimization, authz, rate limits, version behavior, and webhook signature checks.

## 10) Configuration Security

Required:
- Allow only necessary HTTP methods.
- Enable security headers such as CSP, X-Frame-Options or frame-ancestors, X-Content-Type-Options, Referrer-Policy, and HSTS where applicable.
- Do not commit `.env`, credentials, private keys, production config, or local secret dumps.
- Do not use default usernames, passwords, demo tenants, permissive CORS, debug mode, broad IAM policies, or public buckets in production.
- Separate dev/test/prod configuration and fail closed on missing security-critical settings.

Verification:
- Config scan confirms methods, headers, environment separation, CORS, IAM, storage, and debug settings are enforced in all environments.

## 11) Dependencies and Supply Chain

Required:
- Use package manager lockfiles for deterministic dependency trees.
- Run SCA and generate or maintain an SBOM where required.
- Use trusted and maintained libraries only, with approval before adopting risky or security-sensitive dependencies.
- Enforce pre-commit or CI hooks for SAST and SCA where project policy requires them.
- Apply approved security patches promptly.
- Do not use end-of-life libraries, unpinned install scripts, unsafe postinstall behavior, or unreviewed vendored code.

Verification:
- CI gate fails on critical vulnerabilities, missing lockfile/SBOM, known malicious packages, or unsupported runtime/dependency versions.

## 12) AI-Assisted Coding Security

Required when AI coding tools are used:
- Never share source code, customer data, internal configs, secrets, or API keys with unapproved public AI tools.
- Use only approved AI tools with appropriate data-protection settings.
- Developers remain responsible for AI-generated code correctness, licensing, and security.
- AI-generated code must pass SAST/SCA and secure coding checks before commit.
- Explicitly verify high-risk classes: hardcoded secrets, SQL injection, XSS, unsafe deserialization, command injection, path traversal, authz bypass, tenant leaks, and insecure crypto.
- Validate AI suggestions against official documentation and security best practices.
- Track origin notes for complex/high-impact AI-generated code when required for audit.
- Ensure AI-generated code and snippets do not violate license or copyright constraints.

Verification:
- Review sample commits for AI-origin controls, scan evidence, and human verification of high-risk generated code.

## 13) AI System Security

Required when shipping AI runtime features:
- Verify data and model integrity with hashes, signatures, provenance, or trusted deployment channels before training/deployment.
- Detect and mitigate data poisoning through anomaly checks and source validation.
- Restrict, encrypt, and control access to training, evaluation, prompt, and retrieval data.
- Rate-limit model APIs to reduce extraction, abuse, and cost-exhaustion risk.
- Sanitize user input and external content against direct and indirect prompt injection, including retrieved context.
- Do not send sensitive data to AI prompts without classification and approval.
- Never rely on system prompts as the only security control.
- Do not hardcode secrets, hidden policy, or business logic in prompts.
- Enforce tenant isolation for AI context, vector stores, tools, memory, logs, and sessions.
- Validate and filter AI outputs before rendering, execution, query building, workflow decisions, or downstream use.
- Never pass AI output directly into SQL, shell, code execution, browser automation, emails, payments, or destructive operations.
- Require human confirmation for irreversible, external, privileged, or high-risk actions.
- Handle AI timeout, refusal, hallucination, malformed, or invalid outputs safely without exposing internal errors.
- If tool-calling agents are enabled: enforce least privilege, allowlisted tools, confirmation on real-world actions, max-step limits, and audit logs.
- Log and monitor model behavior anomalies while excluding sensitive data and defining retention.
- Reassess risk whenever model, provider, retrieval source, tool set, or prompt policy changes.

Verification:
- Red-team prompt-injection, data-exfiltration, tenant-isolation, tool-abuse, and unsafe-output test cases are included.

## 14) Mobile Application Security

Required when Android/iOS mobile apps exist:
- Detect root on Android or jailbreak on iOS and apply a risk response such as warn, restrict, or deny.
- Detect emulator execution and apply appropriate mitigations for sensitive operations.
- Implement app integrity verification to detect tampering and repackaging.
- Run runtime security checks at startup, resume, and before sensitive operations.
- Enable release obfuscation.
- Sign release APK/IPA with production keystore/certificate only; never use debug keys.
- Store credentials and tokens only in OS secure storage, and minimize local sensitive data retention.

Verification:
- Release checklist includes integrity, signing, obfuscation, secure storage, runtime-defense checks, and sensitive-flow testing.
