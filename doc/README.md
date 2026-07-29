# FedCM Demo — Documentation

This `doc/` folder is a learning-oriented walkthrough of how [FedCM](https://developers.google.com/privacy-sandbox/fedcm)
(Federated Credential Management) actually works, written against this repo's own code so every
concept has a concrete implementation to point at. Read them in order the first time through:

1. **[01-fedcm-concepts.md](01-fedcm-concepts.md)** — what FedCM is, why it exists, and the
   vocabulary (RP, IdP, mediation, Login Status API, disconnect) used everywhere else in these docs.
2. **[02-idp-implementation.md](02-idp-implementation.md)** — every endpoint an Identity Provider
   must expose, what each one is for, and where it's implemented in `internal/idp`.
3. **[03-sp-implementation.md](03-sp-implementation.md)** — everything a Relying Party's frontend
   and backend need to do, and where it's implemented in `internal/sp`.
4. **[04-sequence-diagrams.md](04-sequence-diagrams.md)** — sequence diagrams for the four flows
   this demo exercises: first-time sign-in, silent auto re-authentication, disconnect, and login.

For "how do I run this," see the top-level [README.md](../README.md) instead — this folder is
about *why* it's built this way, not setup steps.

## The one-paragraph version

FedCM lets a Relying Party (RP) ask the browser for a user's identity from an Identity Provider
(IdP) *without* the RP and IdP ever directly talking to each other in JavaScript, and without
third-party cookies. The RP calls `navigator.credentials.get()`; the browser itself fetches a set
of well-known JSON endpoints from the IdP (using the IdP's first-party cookies, which it's allowed
to read because *it's* the browser, not the RP's script), shows a native account-chooser UI, and
hands the RP back a token. Every endpoint below exists to make that browser-mediated handshake
possible without leaking data either side shouldn't see.
