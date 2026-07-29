# 5. FedCM vs. OAuth/OIDC: What It Replaces, and What It Doesn't

Everything so far describes FedCM as if it were a self-contained federated-login protocol, and
inside this demo, it is: there's nothing else in the picture. But go looking for FedCM in the
wild, and the identity/security engineering conversation around it usually converges on one blunt
point: **FedCM is not an extension or replacement of OAuth/OIDC.** It solves a narrower, different
problem. Knowing exactly where the boundary sits is what stops you from reaching for it as a
drop-in "sign in with" replacement, which it isn't.

## What FedCM was actually built to fix

FedCM's motivating problem is the browser-wide removal of third-party cookies, which quietly broke
a handful of specific patterns that older federated-identity products depended on. Not "federated
login" in general, but specifically the parts that leaned on an IdP iframe embedded in the RP's
page having access to the IdP's own cookie:

- **OIDC front-channel logout**: the spec's logout mechanism embeds RP iframes that rely on
  reading the RP's session cookie from a third-party context.
- **Social widgets**: an IdP's "share" or "like" button embedded on an RP page needs the IdP's
  cookie inside that RP's origin to know who's asking.
- **Personalized sign-in buttons**: showing "Continue as Alice" instead of a generic "Sign in"
  button, before the user has done anything, again means reading the IdP's cookie from inside the
  RP's page.
- **Silent session refresh** without a top-level redirect or popup.

Each of these is a case where an IdP's iframe needed third-party-cookie access just to answer "is
this browser already known to you," not to run a full authorization flow. FedCM replaces *that
narrow need* with a browser-mediated API, so answering "who's signed in" no longer needs a cookie
the RP's page can read (or a tracking-adjacent iframe at all). It was never scoped to replace the
authorization side of OAuth/OIDC; see the [FedCM explainer](https://github.com/fedidcg/FedCM/blob/main/explainer.md)
for the working group's own framing of this.

## The gap: what OAuth/OIDC gives you that FedCM has no equivalent for

| OAuth/OIDC concept | FedCM equivalent |
|---|---|
| Authorization code + backend token exchange | None: `id_assertion_endpoint` returns a token directly to the browser, in one shot |
| PKCE / `state` parameter (binds the token response to *this specific* request) | None: nothing in FedCM ties an assertion response back to a specific `navigator.credentials.get()` call beyond the `nonce` the RP itself chooses to check |
| Refresh tokens | None: every re-authentication is a fresh `get()` call |
| Scoped access tokens for calling back into the IdP's APIs | None: FedCM only produces an identity token, not an authorization grant for anything else |
| Standardized revocation/introspection endpoints | Only `disconnect_endpoint`, and it revokes the FedCM *grant*, not any access token a real OIDC flow might have issued separately |

This repo's own shortcuts sit right in that gap: `internal/jwtutil` mints a bare JWT straight from
`id_assertion_endpoint` (see [02-idp-implementation.md §2.5](02-idp-implementation.md)), and the
SP's nonce is generated but never checked against a remembered value (see the "Known demo-only
shortcuts" note in the top-level [README](../README.md)). Neither of those is just demo laziness.
They're the actual shape of the problem every real FedCM integration has to solve on its own,
because FedCM doesn't hand you a standard answer the way OAuth's code-exchange step does.

## The dilemma this creates for a real IdP

Once an IdP has to decide what `id_assertion_endpoint` actually returns, there are really only two
options, and neither maps cleanly onto existing OAuth/OIDC infrastructure.

**Option one: return a bare ID token directly** (what this demo does). This is shaped exactly like
OAuth's *implicit grant*, a token handed straight to the browser with no backend exchange step.
That's the flow the OAuth/OIDC ecosystem has spent years steering people *away* from, precisely
because a token sitting in front-channel JS is more exposed than one only ever seen by a backend
over a server-to-server call.

**Option two: return an opaque code and make the RP's backend exchange it** for the real token,
mirroring OAuth's authorization-code flow. This sounds safer on paper. But FedCM has no
PKCE/`state`-equivalent mechanism to bind that exchange to the specific request that produced the
code, so the RP backend has no standard way to prove "the code I'm redeeming came from the same
context as the `get()` call I initiated," short of inventing its own nonce-based scheme on top.

Either path means building custom protection against the exact class of attacks (replay, request
substitution) that OAuth/OIDC already solved with PKCE and `state`. FedCM just doesn't give you
those primitives, so you end up rebuilding a smaller version of them yourself, per IdP.

## So when is it actually worth adopting?

The honest answer, echoed across most real-world FedCM evaluations: **if your RP and IdP can share
a cookie** (same site, or you control both and can put them on a shared parent domain), **FedCM
buys you very little.** You can solve "is this user signed in" by just reading the cookie, no
endpoint contract from [02](02-idp-implementation.md)/[03](03-sp-implementation.md) needed.

FedCM earns its complexity for **genuinely third-party** relationships, where a first-party cookie
was never on the table in the first place, and where the RP only needs a lightweight identity
signal rather than deep access to the IdP's resources. Think "let this one-off tool confirm the
visitor is a verified account holder," not "give this app ongoing API access on the user's behalf."
For that narrower case, trading a third-party-cookie-dependent iframe for a native,
consent-mediated browser prompt is a real improvement. It's just a smaller slice of "federated
login" than OAuth/OIDC covers as a whole.
