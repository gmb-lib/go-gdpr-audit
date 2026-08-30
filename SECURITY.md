# Security policy

This library records who accessed whose personal data, when, why, and on what lawful basis. That
record is what demonstrates accountability under the GDPR, what a data-subject access request is
answered from, and what a personal-data breach is detected in. It has an unusual property for a
logging client: on the privileged path it is **fail-closed by contract** — if the record cannot be
delivered, the call returns an error so the operation can abort. Code above it relies on that.

It also carries a second, opposite duty. A log about personal data must not itself become a store
of personal data.

Please report security problems privately. Do not open a public issue, pull request or discussion
for anything that could be exploited before a fix exists.

## How to report

Use **[private vulnerability reporting](https://github.com/gmb-lib/go-gdpr-audit/security/advisories/new)**
on this repository. The report stays visible only to you and the maintainers until an advisory is
published, and it gives us one place to discuss and co-ordinate a fix with you.

Please include, as far as you can establish it:

- what the problem is, and what an attacker or an unintended reader gains from it;
- the smallest set of steps that reproduces it, and against which version or commit;
- the configuration it needs — timeouts, breaker settings, which helper — if it only appears under
  particular settings;
- whether you have told anyone else, and whether a disclosure date already binds you.

**Please do not send us anyone's personal data.** If a finding is about a leak, describe the field
and its shape; placeholder values are enough.

## What happens next

- We acknowledge a report within **five working days**.
- We tell you whether we can reproduce it, and what we think its severity is, as soon as we know.
- We keep you updated while a fix is prepared, and we agree a disclosure date with you. Our default
  is to publish an advisory once a fix is available, and in any case within **90 days** of the
  report — earlier if the problem is already public or being exploited.
- We credit you in the advisory unless you would rather stay anonymous.

There is no bug-bounty programme. We are grateful anyway, and we say so publicly.

## What we consider most serious

- A **privileged** record that fails to deliver while the call returns success. Fail-closed is this
  library's contract on that path, and a silent success lets an elevated access — an export, an
  erasure, an operator read — happen with nothing recording it.
- A record lost with nothing noticing: dropped by the circuit breaker, the outbox or a drain retry
  without reaching the dead-letter sink.
- Personal data in the record itself — a national identifier, name or e-mail address in
  `DataSubjects` where a pseudonymous internal reference is required, or an attribute carrying the
  very data the record is about.
- A record attributed to the wrong actor, subject, purpose or lawful basis. A data-subject access
  request answered from this log then answers wrongly, which is a compliance failure and a privacy
  failure at once.
- The outbox persisting records somewhere an unauthorised reader can reach them, keeping them
  longer than intended, or writing them with permissions wider than the process needs.
- The injected poster reaching a destination other than the configured sink, or a post that escapes
  its configured timeout.
- A routine record that should have been privileged, or a helper whose delivery policy does not
  match the one documented for it — because callers choose the helper on the strength of that
  policy.

Denial of service and findings that need an already-compromised host are in scope but lower
priority. Reports about outdated dependencies are welcome where you can show the vulnerable path
is actually reachable.

## Scope

This policy covers the code in this repository. It does not cover the access-audit service that
stores the records, its database, the transport the injected poster is built on, or the services
that import this library — report those to the parties that operate them.

## Supported versions

Security fixes land on the most recent release. Older tags are not patched; if you are pinned to
one, the fix is to move forward. This module is pinned in lockstep with the platform kit, so a fix
may require moving that pin too.
