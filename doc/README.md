# FedCM 101: Documentation

This `doc/` folder is walkthrough of how [FedCM](https://developers.google.com/privacy-sandbox/fedcm)
(Federated Credential Management) actually works.

1. **[01-fedcm-concepts.md](01-fedcm-concepts.md)**: what FedCM is, why it exists, and the
   vocabulary (RP, IdP, mediation, Login Status API, disconnect) we use everywhere else in these
   docs.
2. **[02-idp-implementation.md](02-idp-implementation.md)**: every endpoint an Identity Provider
   has to expose, what each one is for, and where it's implemented in `internal/idp`.
3. **[03-sp-implementation.md](03-sp-implementation.md)**: everything a Relying Party's frontend
   and backend need to do, and where it's implemented in `internal/sp`.
4. **[04-sequence-diagrams.md](04-sequence-diagrams.md)**: sequence diagrams for the four flows
   this demo exercises: first-time sign-in, silent auto re-authentication, disconnect, and login.
5. **[05-fedcm-vs-oauth-oidc.md](05-fedcm-vs-oauth-oidc.md)**: the problem FedCM actually solves,
   the gap between it and OAuth/OIDC (no PKCE/state equivalent, no code exchange, no refresh
   tokens), and when it's genuinely worth adopting instead of just sharing a cookie.

For "how do I run this," check the top-level [README.md](../README.md) instead.
