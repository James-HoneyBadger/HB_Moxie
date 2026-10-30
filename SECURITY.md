# Security Policy

Honey Ledger stores personal financial records, so security reports are taken
seriously and handled privately.

## Supported Versions

Only the latest published release receives security fixes. Fixes are shipped in
a new patch release; there are no long-term support branches.

| Version | Supported |
|---------|-----------|
| Latest `1.4.x` release | ✅ |
| Older releases | ❌ |

## Verifying Releases

Release artifacts can be checked without trusting this repository's hosting:

**Build provenance.** GitHub attests that each published file was produced by
the release workflow of this repository. Verify any artifact with:

```bash
gh attestation verify <file> --repo James-HoneyBadger/honey-ledger
```

**Detached signature.** Artifacts are signed with
[minisign](https://jedisct1.github.io/minisign/) when signing is configured.
Each release publishes the `minisign.pub` public key next to the installers:

```bash
minisign -V -p minisign.pub -m <file>
```

If a release has no `.minisig` files, signing was not configured for it —
prefer the provenance attestation, or build from source.

## Reporting a Vulnerability

**Do not open a public issue, pull request, or discussion for a security
problem.**

1. **Preferred:** open a private advisory at
   <https://github.com/James-HoneyBadger/honey-ledger/security/advisories/new>.
   GitHub private vulnerability reporting is enabled for this repository.
2. **Alternative:** email <james@honey-badger.org> with a description of the
   issue and reproduction steps.

Please include the affected version, platform, and — where possible — a minimal
reproduction. If you plan to publish a write-up, let us know so a fix can ship
first.

### What to expect

| Stage | Target |
|-------|--------|
| Acknowledgement | within 3 business days |
| Assessment and severity rating | within 10 business days |
| Fix or documented mitigation | within 30 days for high and critical issues |
| Public disclosure | coordinated with the reporter after a release is available |

Reporters are credited in the release notes unless they ask to stay anonymous.

## Scope

### In scope

- Authentication and session handling: JWT issuance, token revocation, API and
  MCP token scoping, two-factor authentication, external OAuth2/OIDC login
- Authorization: cross-user data access, feature restrictions, admin-only paths
- Injection: SQL and ORM query construction, template rendering, command
  execution
- File handling: transaction picture uploads, avatar handling, import parsing,
  backup archive extraction (including path traversal and zip-slip)
- Desktop shell: renderer-to-main boundaries, `app://` protocol handler, IPC,
  navigation and window-open restrictions
- Local attack surface: DNS rebinding or CSRF against the loopback HTTP server
- Denial of service caused by malformed input in import parsers

### Out of scope

- Physical access to an unlocked machine, or a compromised OS user account
  (the local user is the trust boundary)
- The database being unencrypted at rest — this is a known, documented
  limitation rather than a vulnerability
- Automatic login in single-user desktop mode, which is an intentional feature;
  it requires exactly one account and is restricted to loopback clients
- Vulnerabilities in third-party services the user configures (mail, LLM
  providers, map tiles, MinIO, WebDAV)
- Reports produced solely by automated scanners without a demonstrated
  exploitation path

## Threat Model

Honey Ledger is a single-user desktop application:

- The Go backend binds to loopback only and refuses non-loopback bind
  addresses.
- Each session token carries its own random signing secret stored in the
  database, so revocation is immediate.
- The Electron renderer runs sandboxed with context isolation, no Node
  integration, a restrictive Content-Security-Policy, and all permission
  requests denied.
- Other local processes are **not** trusted; the loopback server must not be
  reachable or controllable from a web page or from other users on the host.

## Safe Harbor

Security research conducted in good faith against your own installation, using
data you own, will not result in legal action. Please avoid privacy violations,
data destruction, and service disruption while testing.
