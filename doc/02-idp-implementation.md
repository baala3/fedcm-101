# 2. Building the Identity Provider

Everything here lives in `internal/idp`. An IdP is really just a normal web app (login form,
sessions, a user table) plus a fixed contract of extra JSON endpoints the browser calls on the
RP's behalf. This doc walks through that contract in the order the browser actually calls it.

> Reference: [Chrome for Developers — Implement the identity provider](https://developer.chrome.com/docs/identity/fedcm/implement/identity-provider)

## 2.1 `/.well-known/web-identity`: proving ownership of the config URL

**Why it exists:** an RP could point `navigator.credentials.get()` at *any* configURL, including
one it doesn't own, to trick the browser into treating some other site as an IdP. The well-known
file, fetched from the IdP's own origin without credentials and which must not redirect, is the IdP
explicitly vouching "yes, this config URL is mine."

**Implementation:** `internal/idp/handlers_wellknown.go`, `handleWellKnown`. Newer Chrome/Edge also
require `accounts_endpoint` and `login_url` to be pinned here directly (not just inside the config
JSON) whenever the IdP also serves `client_metadata_endpoint` (see [§1](01-fedcm-concepts.md) for
why). Response:

```json
{
  "provider_urls": ["http://localhost:8080/fedcm.json"],
  "accounts_endpoint": "http://localhost:8080/fedcm/accounts",
  "login_url": "http://localhost:8080/login"
}
```

## 2.2 `/fedcm.json`: the config document

**Why it exists:** a single place listing every other endpoint, plus the branding shown in the
account chooser (background color, icon).

**Implementation:** `internal/idp/handlers_config.go`, `handleConfig`. This is fetched without
credentials. We still defensively check `Sec-Fetch-Dest: webidentity` (the header Chrome/Edge
attach to every FedCM request), even though nothing else could realistically be asking for this
file. Response:

```json
{
  "accounts_endpoint": "http://localhost:8080/fedcm/accounts",
  "client_metadata_endpoint": "http://localhost:8080/fedcm/client_metadata",
  "id_assertion_endpoint": "http://localhost:8080/fedcm/assertion",
  "disconnect_endpoint": "http://localhost:8080/fedcm/disconnect",
  "login_url": "http://localhost:8080/login",
  "branding": { "background_color": "#1a73e8", "color": "#ffffff", "icons": [...] }
}
```

## 2.3 `/fedcm/client_metadata`: per-RP disclosure links

**Why it exists:** the account chooser shows a privacy policy / terms of service link so the user
knows what they're agreeing to share data with. It's fetched uncredentialed, but *with* knowledge
of which RP (`client_id`) is asking, which is exactly why §2.1's stricter well-known requirement
exists: to stop this RP-awareness from being (ab)used to vary account data per RP undetectably.

**Implementation:** `internal/idp/handlers_client_metadata.go`. It needs
`Access-Control-Allow-Origin` (no credentials involved, so no `Allow-Credentials` needed), because
unlike the other FedCM endpoints, this fetch is subject to normal CORS. More on that in §2.8.

## 2.4 `/fedcm/accounts`: who's signed in, credentialed

**Why it exists:** this is the one request where the browser actually attaches the IdP's session
cookie, so the IdP can answer "which account(s), if any, does this browser currently have a
session for." It's deliberately **not** RP-aware (see §1); the browser cross-references the
response against the requesting RP's `client_id` itself.

**Implementation:** `internal/idp/handlers_accounts.go`, `handleAccounts`.

- No session cookie → **401**. This matters more than it looks: it's what makes the Login Status
  API's mismatch handling work (§2.6). If the IdP claimed "logged-in" but accounts says otherwise,
  the browser corrects itself.
- With a session → returns the account, including `approved_clients`: every `client_id` this user
  has an active grant for (from the `grants` sqlite table, `internal/idp/store.go`,
  `ClientsGrantedFor`). This is what lets a *returning* user skip the "Continue as ___" consent
  screen, since the browser checks whether the requesting RP is already in that list.

```json
{ "accounts": [{ "id": "1", "name": "Alice Adams", "email": "...", "approved_clients": ["http://localhost:8081"] }] }
```

## 2.5 `/fedcm/assertion`: minting the token

**Why it exists:** once the user picks an account in the native chooser, the browser POSTs here
(credentialed, form-encoded) to actually get a token to hand back to the RP.

**Implementation:** `internal/idp/handlers_assertion.go`, `handleAssertion`.

1. Validate `Sec-Fetch-Dest: webidentity` and that `Origin` matches the one RP this demo knows
   (`RPOrigin` in `internal/idp/server.go`). A real IdP would look this up per `client_id` instead
   of hardcoding a single origin.
2. Re-check the session cookie (the user still has to be signed in).
3. Read `account_id`, `client_id`, `nonce` from the POST form and cross-check `account_id` against
   the signed-in user, so a client can't assert a token for an account it didn't actually pick.
4. Mint a JWT via `internal/jwtutil/token.go` (`Mint`): `iss` is the IdP origin, `sub` is the
   account id, `aud` is the `client_id`, plus profile claims and the `nonce` (carried in the JWT
   `jti` field, so a replay-detection scheme could use it, though this demo doesn't check it
   against anything stored (see the limitations note in the top-level README).
5. Record a `grants` row (`store.AddGrant`) so future `accounts` responses include this RP in
   `approved_clients`.
6. Return `{"token": "<jwt>"}`.

Error responses use the FedCM-specified shape so the browser can show something meaningful:
`{"error": {"code": "access_denied", "url": ""}}` (see `writeAssertionError`).

## 2.6 The login page and the Login Status API

**Why it exists:** the browser needs to know, before it even tries the accounts endpoint, whether
this IdP thinks anyone is signed in. Otherwise every page load on every site would need to
speculatively fetch every known IdP's accounts endpoint. The **Login Status API** is the IdP
proactively telling the browser "a user just signed in / signed out here," via the `Set-Login`
response header.

**Implementation:** `internal/idp/handlers_auth.go`.

- `handleLoginPage` (`GET /login`): a normal HTML login form. This is what the browser opens (in
  a popup/tab) when it wants a user to sign in but has no session for them yet. It's driven by
  `login_url` from the config, not anything FedCM-specific in the page itself.
- `handleLoginSubmit` (`POST /login`): on success, creates an in-memory session
  (`internal/idp/session.go`), sets the `idp_session` cookie, and sets `Set-Login: logged-in`. The
  response body is a tiny page whose script calls `IdentityProvider.close()`, which tells the
  browser "the popup's done its job, close it and retry the accounts fetch automatically."
  `window.IdentityProvider` actually exists on *every* page, not just inside that special popup, so
  `close()` silently no-ops if this page was loaded as a normal top-level navigation instead. The
  script also arms a fallback `setTimeout` redirect to `/`, just in case `close()` doesn't actually
  tear the page down.
- `handleLogout` (`POST /logout`): clears the session and sends `Set-Login: logged-out`.

## 2.7 `/fedcm/disconnect`: RP-initiated revocation

**Why it exists:** lets an RP page call `IdentityCredential.disconnect()` to explicitly unlink
itself from a user's IdP account (say, an "unlink your Google account" button), without needing a
whole separate settings UI on the IdP side.

**Implementation:** `internal/idp/handlers_disconnect.go`, `handleDisconnect`. Same
Origin/Sec-Fetch-Dest checks as assertion, then `store.RemoveGrant` deletes the `grants` row and
returns `{"account_id": "..."}`.

There's a second, IdP-side way to reach the same table: the account home page
(`internal/idp/handlers_home.go`, `handleHome` / `handleRevoke`, `internal/idp/templates/home.html`)
lists every RP a user has granted, with its own "Revoke" button. That's plain form-based
self-service revocation, not part of the FedCM API at all. It's just a normal use of the same
`grants` table from the IdP's own UI.

## 2.8 CORS: the part that's easy to get wrong

Even though the credentialed FedCM endpoints (`accounts`, `assertion`, `disconnect`) and the client
metadata endpoint are invoked internally by the browser rather than by a page's `fetch()`, they're
still subject to real CORS checks in current Chrome/Edge. Every one of them, **including error
responses**, needs:

```
Access-Control-Allow-Origin: http://localhost:8081   (the exact RP origin, not *)
Access-Control-Allow-Credentials: true               (only on the credentialed ones)
```

See `setCORSHeaders` in `internal/idp/server.go`, called at the very top of each handler before any
early-return error path. A response missing these headers surfaces to the RP page as a generic
`ERR_FAILED` / "Server did not send the correct CORS headers" network error, which gives no hint
about what actually went wrong server-side. That vagueness is exactly what makes this easy to miss
while debugging.
