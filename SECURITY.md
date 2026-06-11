# Security Policy

Ghost executes commands and edits files on the user's machine, and connects to
third-party AI providers. We take the security of the tool and its users
seriously.

## Supported versions

Security fixes are applied to the latest released version. Please upgrade to the
newest release before reporting an issue.

## Reporting a vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, report them privately using GitHub's
[private vulnerability reporting](https://github.com/swadhinbiswas/Ghost/security/advisories/new)
(Security → Advisories → "Report a vulnerability").

When reporting, please include:

- A description of the vulnerability and its impact
- Steps to reproduce (proof-of-concept if possible)
- Affected version(s) and platform
- Any suggested remediation

We aim to acknowledge reports within **72 hours** and to provide a remediation
timeline after triage.

## Scope

Examples of issues we consider in scope:

- Command execution escaping the intended permission/approval flow
- Leakage of API keys, OAuth tokens, or other secrets (e.g., to logs)
- Path traversal or unintended file writes outside the working directory
- Sandbox escape from the Docker-backed shell
- Supply-chain risks in the build or release pipeline

## Handling secrets

Ghost reads credentials such as `NVIDIA_API_KEY` and provider OAuth tokens.
These must never be written to logs, committed to the repository, or transmitted
to anywhere other than the configured provider endpoint. If you find a case
where a secret is exposed, treat it as a security issue and report it privately.

## Safe usage for users

- Review permission prompts before approving file edits or shell commands.
- Prefer running untrusted tasks with the Docker sandbox enabled when available.
- Keep your Ghost installation up to date.
