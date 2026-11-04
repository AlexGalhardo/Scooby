# SDD-01 — Google identity and dashboard

> Current implementation and explicit deviations are in the [ROADMAP](../../ROADMAP.md)
> and [ADRs 0011–0013](../adrs/README.md). Local technical acceptance is distinct
> from real OAuth/broker, Railway publication and prolonged financial observation.

Status: RF-01–RF-04 and personal use confirmed. Admission restricted to one account confirmed in Q6; recovery still pending.

## Navigation contract

| Route | Visitor | Authenticated |
| --- | --- | --- |
| GET `/` | 302 to `/login`; Inertia uses a compatible navigation redirect | Renders `Dashboard` |
| GET `/login` | Renders `Auth/Login` with one Google button | Redirects to `/` |
| GET `/auth/google/redirect` | Starts OAuth/OIDC with state, nonce and PKCE | May return to `/` |
| GET `/auth/google/callback` | Validates the flow and authorization; creates session and goes to `/` | Handle invalid flow without silently switching identity |
| POST `/logout` | Does not create a session | CSRF, invalidates session and redirects to `/login` |
| GET `/dashboard` | Same guard | Optional alias that redirects to `/` |
| GET `/up` | Minimal liveness | Same behavior |

`/posts` and demo forms will be removed or protected during replacement.
Local password, sign-up, reset and email verification routes must not remain
accessible merely because the `install:auth --social google` generator wrote them.
Register only Google, with no arbitrary provider chosen by the client.

## Identity and session

Use the Vukem `social.Google` support; do not implement manual OAuth without reason.
Validate OIDC signature and claims, issuer, audience, expiration, nonce, state and
single use. Stable identity is provider + subject, not the editable email. Verify
the email when used in the policy. Do not link existing accounts by email without an
explicit policy; do not accept a self-declared address as authorization.

Rotate the session on login, use appropriate Secure/HttpOnly/SameSite cookies, TTL and
revocation. Single-use OAuth state must be shared between instances.
Use persistent cache/session compatible with Vukem and migrations. An expired session
in JSON/stream returns 401; HTML navigation leads to `/login`. CSRF on mutations.
Limit attempts and return messages without codes/tokens or internal details.

As Q1 confirmed personal use, the scaffold's implicit sign-up needs a gate: only pre-authorized identity may enter; controlled bootstrap of the owner and recovery
follow Q6: verified email configured on the server, bound to the stable subject;
recovery will still be detailed. SaaS is outside the MVP. A valid Google account does not imply permitted access.

## Interface

Use SidebarProvider/Sidebar/SidebarInset and shadcn/ui components for cards,
tables, dialogs, forms and feedback. Proposed navigation: Overview,
Market, Strategies, Backtests, Orders, Positions, Risk, Audit and Settings.
Show only delivered modules or a clear unavailable state.

Header: account, market, PAPER/REAL, connectivity and data delay. Cards:
equity, realized/unrealized P&L, exposure and drawdown. Content: chart,
recent orders and engine health. Currency, period, source and update time
are explicit. Missing P&L does not become zero; simulated does not appear as real.

Minimal props only: public user, permissions and authorized snapshots. Never
credentials, Google tokens or broker secrets. Official calculations in the backend.
Theme, language, default period and module positions await a UX round.

## Acceptance

Meet the auth/UI tests in `../validation.md`. Logging in with Google returns to `/`;
a visitor does not see the dashboard before the redirect; logout revokes the session; the single button
does not introduce password login. SSR is not a confirmed requirement. Responsive and
accessible sidebar; risk actions are authorized on the server and audited.

## Verified notes for implementation

In the existing frontend, the `@` alias must point to `resources/js`, and the CSS lives in
`resources/css/app.css`; do not recreate a parallel `src/` project. The current
shadcn/ui documentation has variants of primitives: choose one (Base UI, Radix or another
supported one) and keep consistency before copying components. This choice is still a
proposed technical detail. On Google, limit scopes to `openid email profile`.

## Single-account bootstrap — confirmed in Q6

Initial configuration `OWNER_GOOGLE_EMAIL` (proposed name) contains the authorized email,
normalized according to a restricted contract, without Gmail dot/alias heuristics.
The callback requires a verified email and comparison with the configuration before the session
is created. The first binding writes the subject with a transaction and uniqueness, preventing
two concurrent callbacks from registering different owners. After binding,
the subject is the access key; an email change does not register a new owner.
The policy for updating/recovering the binding will be chosen before deploy.
No public screen requests the owner configuration. Secrets and the owner's real email
will be configured in the environment, not in this document.
