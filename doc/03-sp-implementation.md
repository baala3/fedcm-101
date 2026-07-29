# 3. Building the Relying Party

Everything here lives in `internal/sp`. Unlike the IdP, the RP has no FedCM-specific *server*
contract to satisfy. From the IdP's perspective the RP is just "some `client_id`." All the FedCM
work on this side happens in the browser JS; the backend's job is limited to verifying whatever
token that JS hands it, then running an ordinary cookie session.

> Reference: [Chrome for Developers — Implement the relying party](https://developer.chrome.com/docs/identity/fedcm/implement/relying-party)

## 3.1 Requesting a credential

**Implementation:** `internal/sp/static/main.js`, `requestCredential()` / `providerConfig()`.

```js
navigator.credentials.get({
  identity: {
    providers: [{
      configURL: "http://localhost:8080/fedcm.json",
      clientId: "http://localhost:8081",
      params: { nonce: crypto.randomUUID() },
    }],
  },
  mediation, // "optional" or "silent" — see below
});
```

- `configURL` / `clientId` are the only two things the RP needs to know about the IdP up front.
  Everything else (which endpoints exist, branding, etc.) comes from the IdP's own config document.
- `nonce` is generated fresh per attempt and round-trips through the minted JWT
  (`internal/jwtutil`), for replay binding. It goes inside `params`, not as a top-level field.
  Newer Chrome/Edge deprecated the top-level form, so don't copy older examples that use it.
- This call is what kicks off the entire browser-mediated sequence described in
  [01-fedcm-concepts.md](01-fedcm-concepts.md): well-known → config → accounts → (login popup, if
  needed) → account chooser UI → assertion → resolved promise.

**The one-outstanding-request rule:** the browser only allows one `navigator.credentials.get()`
call in flight at a time *per page*. This demo tries a silent call on page load (§3.2) and a
manual call on button click, and if those ever overlap, the second one throws
`NotAllowedError: Only one navigator.credentials.get request may be outstanding at one time`. We
handle this in `requestCredential()` with an `AbortController`: starting a new request aborts
whichever one was already running, instead of leaving both racing each other.

## 3.2 Silent auto re-authentication

**Implementation:** `attemptSilentSignIn()` in `main.js`, called on the home page's
`DOMContentLoaded`.

`mediation: "silent"` asks the browser to sign the user back in **with no UI at all**. It's what
gives you "you were signed in last time, so just resume" instead of a jarring chooser popup on
every visit. It only succeeds if:

- the user has an existing grant for this RP (`approved_clients` included this `client_id` the
  last time `/fedcm/accounts` was fetched), **and**
- there's exactly one matching account, **and**
- the browser isn't in the post-sign-in **quiet period** (an anti-abuse cooldown, see §1). This is
  why silent auto-reauth can look like it works sometimes and not others right after you've just
  signed in or disconnected. That's the browser throttling it, not a bug in this code.

When any of those conditions fail, the promise rejects and we just leave the "Sign in" button
visible. There's no special handling needed beyond catching the rejection, which is also why the UI
shows a transient "Checking for an existing sign-in…" status that clears itself, rather than
showing an error.

## 3.3 Exchanging the token for a session

**Implementation:** `completeSignIn()` in `main.js`, posting to `internal/sp/handlers_session.go`
(`handleCreateSession`, `POST /session`).

Once `navigator.credentials.get()` resolves, the RP has an `IdentityCredential` object whose
`.token` is whatever the IdP's assertion endpoint returned (a JWT, in this demo). The RP's own JS
then does a plain same-origin `fetch("/session", { body: { token } })`. Server-side:

1. `jwtutil.Verify(token, SPOrigin)` checks the HMAC signature and that the token's `aud` claim
   matches this RP's own origin, so a token minted for some *other* `client_id` can't be replayed
   here.
2. On success, sets an `sp_session` cookie holding the raw JWT (re-verified on every subsequent
   request in `internal/sp/session.go`, `currentClaims()`), so an expired token stops being treated
   as a valid session automatically. No separate expiry bookkeeping needed.

From here on, `/profile` (`internal/sp/handlers_pages.go`) is just an ordinary cookie-authenticated
page. FedCM's job ended the moment the token was verified.

## 3.4 Disconnecting

**Implementation:** `disconnect()` in `main.js`, calling the static
`IdentityCredential.disconnect({ configURL, clientId })` method. Note this isn't
`navigator.credentials.*`, it's a different global. The browser turns this into a credentialed
POST to the IdP's `disconnect_endpoint`. Once that resolves, the RP's JS calls its own
`POST /disconnected` (`internal/sp/handlers_session.go`, `handleDisconnected`) to drop the local
`sp_session` cookie too. The IdP call only revokes the *grant*; clearing the RP's own session is
the RP's job.

## 3.5 What the RP deliberately does *not* need

- No CORS setup of its own. `/session` and `/disconnected` are same-origin fetches from the page's
  own JS to its own backend.
- No knowledge of the IdP's session cookie, user table, or login flow. All of that stays opaque
  behind `navigator.credentials.get()`.
- No polling or webhook to find out about revocation. If a user revokes the RP's access from the
  IdP's own account page (`internal/idp/handlers_home.go`, a plain HTML form outside the FedCM
  API), the RP simply won't get a token next time it asks.
