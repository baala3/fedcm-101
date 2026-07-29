# 5. FedCM vs. OAuth/OIDC:

**FedCM is not a replacement for OAuth/OIDC.** It only solves the browser privacy issues caused by the removal of third-party cookies.

FedCM replaces a few cookie-dependent patterns, such as:
- OIDC front-channel logout
- Social widgets (e.g., Like/Share buttons)
- Personalized sign-in buttons ("Continue as Alice")
- Silent session refresh

Instead of letting an IdP iframe read third-party cookies, the browser provides this information through a dedicated API. OAuth/OIDC authorization flows remain unchanged, FedCM only replaces these cookie-dependent interactions, not the authentication or authorization protocols themselves.

https://github.com/w3c-fedid/FedCM


## The gap: what OAuth/OIDC gives you that FedCM has no equivalent for

| OAuth/OIDC concept | FedCM equivalent |
|---|---|
| Authorization code + backend token exchange | None: `id_assertion_endpoint` returns a token directly to the browser, in one shot |
| PKCE / `state` parameter (binds the token response to *this specific* request) | None: nothing in FedCM ties an assertion response back to a specific `navigator.credentials.get()` call beyond the `nonce` the RP itself chooses to check |
| Refresh tokens | None: every re-authentication is a fresh `get()` call |
| Scoped access tokens for calling back into the IdP's APIs | None: FedCM only produces an identity token, not an authorization grant for anything else |
| Standardized revocation/introspection endpoints | Only `disconnect_endpoint`, and it revokes the FedCM *grant*, not any access token a real OIDC flow might have issued separately |

This demo highlights that gap. The `id_assertion_endpoint` returns a JWT directly, and while the SP generates a nonce, it doesn't validate it later. 
These aren't just demo shortcuts—they reflect gaps in FedCM itself. Unlike OAuth's authorization code flow, FedCM doesn't provide standardized mechanisms for request binding or replay protection, so real IdPs must implement them on their own.


## The dilemma this creates for a real IdP

Once an IdP has to decide what `id_assertion_endpoint` actually returns, there are really only two
options, and neither maps cleanly onto existing OAuth/OIDC infrastructure.

**Option one: return a bare ID token directly** (what this demo does). This is shaped exactly like
OAuth's *implicit grant*, a token handed straight to the browser on front channel.
That's the flow the OAuth/OIDC ecosystem has spent years steering people *away* from, and
because a token sitting in front-channel JS is more exposed than one only ever seen by a backend
over a server-to-server call.

**Option two: return an opaque code and make the RP's backend exchange it** for the real token,
mirroring OAuth's authorization-code flow. This sounds safer, but FedCM has no
PKCE/`state` equivalent mechanism to bind that exchange to the specific request that produced the
code, so RP backend has no standard way to prove "the code I'm redeeming came from the same
context as the `get()` call I initiated," short of inventing its own nonce-based scheme on top.

Either path means building custom protection against the exact class of attacks (replay, request
substitution) that OAuth/OIDC already solved with PKCE and `state`. FedCM just doesn't give you
those, so you end up rebuilding a smaller version of them yourself, per IdP.

## So when is it actually worth adopting?

If your RP and IdP can share cookies (same-site or under the same parent domain), 
FedCM provides almost no benefit as you can already determine whether a user is signed in using first-party cookies.

FedCM is most valuable for true third-party identity providers, where third-party cookies are unavailable. It works well when the RP only needs to verify a user's identity, not request long-lived API access. In these cases, FedCM replaces cookie-dependent iframes with a browser-managed sign-in flow, improving privacy without replacing OAuth/OIDC.
