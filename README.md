# FedCM 101

A minimal, fully local [FedCM](https://developers.google.com/privacy-sandbox/3pcd/fedcm) setup
with two independent Go servers on `localhost`:

- **IdP** (`cmd/idp`, `http://localhost:8080`) — the identity provider: login page,
  well-known/config/accounts/assertion/disconnect endpoints, SQLite-backed users + consent grants.
- **SP** (`cmd/sp`, `http://localhost:8081`) — the relying party: a "Sign in" button that calls
  `navigator.credentials.get()`, a profile page, and a disconnect button.

Chrome treats `http://localhost` as a secure context, so no TLS setup is required.

**New to FedCM?** See [`doc/`](doc/README.md) for a concept-by-concept walkthrough of how the
protocol works and how each endpoint here implements it, with sequence diagrams and a log of real
errors hit (and fixed) while building this.

## Requirements

- Go 1.22+ (a `go.mod` toolchain directive will auto-fetch a newer patch release if needed).
- A recent **Chrome** (or other Chromium-based browser) with FedCM enabled. The basic sign-in
  flow ships by default in modern Chrome. `IdentityCredential.disconnect()` and some Login Status
  API nuances are newer additions — if a step below doesn't trigger, check `chrome://flags` for
  FedCM-related flags, or try Chrome Canary.
- Not Incognito/Guest mode — some FedCM behavior is restricted there.

## Running it

In two separate terminals:

```
make run-idp   # http://localhost:8080
make run-sp    # http://localhost:8081
```

The IdP seeds two demo users on first run (`data/idp.db`): `alice` / `bob`, password `password123`.

## Manual walkthrough

1. Open `http://localhost:8081`. You're not signed in anywhere yet, so the silent
   auto-reauth attempt fails quietly and you see the "Sign in with Demo IdP" button.
2. Click it. Chrome shows "Sign in to Demo RP with your IdP account" and offers to open the
   IdP's login page (since you have no IdP session yet).
3. Log in as `alice` / `password123`. The login page calls `IdentityProvider.close()`, which
   closes the popup and lets the browser retry the accounts fetch automatically — the native
   FedCM account chooser should now appear.
4. Pick the account. You land on `/profile` showing the claims decoded from the id token the IdP
   minted (name, email, picture).
5. Click **Disconnect from IdP** — this calls `IdentityCredential.disconnect()`, which hits the
   IdP's disconnect endpoint and revokes the stored consent grant, then clears the SP session.
6. Reload `http://localhost:8081` before disconnecting again vs. after: with an active grant, the
   silent (`mediation: 'silent'`) auto-reauth on page load should sign you back in without any UI;
   right after disconnecting, it should fail and fall back to the button.
7. **Log out (keep consent)** only clears the SP's local session — the IdP grant stays, so silent
   auto-reauth still works afterward, unlike disconnect.

## Sanity-checking endpoints directly

Before touching the browser, you can confirm the IdP's JSON endpoints look right:

```
curl http://localhost:8080/.well-known/web-identity
curl http://localhost:8080/fedcm.json
curl http://localhost:8080/fedcm/client_metadata
curl -i http://localhost:8080/fedcm/accounts        # expect 401, no session cookie
```

## Known demo-only shortcuts

- The IdP signs id tokens with a hardcoded shared HMAC secret (`internal/jwtutil`) instead of an
  asymmetric key + JWKS — fine for a same-machine demo, never do this in production.
- IdP sessions live in memory and reset on restart; only `users` and `grants` persist in SQLite.
- Only one RP origin (`http://localhost:8081`) is recognized — a real IdP would look up allowed
  origins per `client_id`.
- The per-request `nonce` is minted and passed through the id token but not checked against a
  server-stored expected value, since the SP is stateless between page load and sign-in here.
