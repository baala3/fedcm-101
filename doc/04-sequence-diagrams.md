# 4. Sequence Diagrams

These trace the four flows this demo exercises. "Browser (FedCM)" represents the browser's
internal FedCM machinery: the account chooser UI, and the fetches it makes on the RP's behalf. That's
also why the RP and IdP columns never message each other directly.

## 4.1 First-time sign-in (no existing IdP session, no prior grant)

```mermaid
sequenceDiagram
    actor User
    participant SP as SP page JS<br/>(main.js)
    participant FedCM as Browser (FedCM)
    participant IdP as IdP server

    User->>SP: Click "Sign in with Demo IdP"
    SP->>FedCM: navigator.credentials.get({mediation:"optional"})
    FedCM->>IdP: GET /.well-known/web-identity (no creds)
    IdP-->>FedCM: provider_urls, accounts_endpoint, login_url
    FedCM->>IdP: GET /fedcm.json (no creds)
    IdP-->>FedCM: config (accounts/assertion/disconnect/login_url)
    FedCM->>IdP: GET /fedcm/accounts (with IdP cookies)
    IdP-->>FedCM: 401 (no session yet)

    FedCM->>User: "Continue to idp to sign in" prompt
    User->>FedCM: Clicks continue
    FedCM->>IdP: opens login_url in popup (GET /login)
    IdP-->>User: login form
    User->>IdP: POST /login (username, password)
    IdP-->>FedCM: Set-Cookie: idp_session, Set-Login: logged-in<br/>page calls IdentityProvider.close()
    FedCM->>FedCM: closes popup

    FedCM->>IdP: GET /fedcm/accounts (retry, with cookie)
    IdP-->>FedCM: 200 {accounts:[{id, name, email, ...}]}
    FedCM->>IdP: GET /fedcm/client_metadata
    IdP-->>FedCM: privacy_policy_url, terms_of_service_url

    FedCM->>User: native account chooser UI
    User->>FedCM: picks account
    FedCM->>IdP: POST /fedcm/assertion (account_id, client_id, nonce)
    IdP->>IdP: verify origin, mint JWT, store grant
    IdP-->>FedCM: {token: "<jwt>"}
    FedCM-->>SP: credentials.get() resolves {token}

    SP->>SP: POST /session {token} (same-origin)
    SP->>SP: verify JWT, set sp_session cookie
    SP-->>User: redirect to /profile
```

## 4.2 Returning user — silent auto re-authentication

```mermaid
sequenceDiagram
    actor User
    participant SP as SP page JS
    participant FedCM as Browser (FedCM)
    participant IdP as IdP server

    User->>SP: loads http://localhost:8081/
    SP->>FedCM: navigator.credentials.get({mediation:"silent"})
    FedCM->>IdP: GET /fedcm/accounts (IdP cookie still valid)
    IdP-->>FedCM: 200 {accounts:[{..., approved_clients:["...:8081"]}]}

    alt grant present, one match, not in quiet period
        FedCM->>IdP: POST /fedcm/assertion (no UI shown)
        IdP-->>FedCM: {token}
        FedCM-->>SP: resolves {token}
        SP->>SP: POST /session, set cookie
        SP-->>User: redirect to /profile (no click needed)
    else quiet period active / no grant / multiple accounts
        FedCM-->>SP: rejects (NotAllowedError)
        SP-->>User: shows "Sign in with Demo IdP" button
    end
```

## 4.3 Disconnect (RP-initiated revocation)

```mermaid
sequenceDiagram
    actor User
    participant SP as SP page JS
    participant FedCM as Browser (FedCM)
    participant IdP as IdP server

    User->>SP: clicks "Disconnect from IdP" on /profile
    SP->>FedCM: IdentityCredential.disconnect({configURL, clientId})
    FedCM->>IdP: POST /fedcm/disconnect (client_id) — credentialed
    IdP->>IdP: delete grants row for (user, client_id)
    IdP-->>FedCM: {account_id}
    FedCM-->>SP: disconnect() resolves

    SP->>SP: POST /disconnected (clears sp_session cookie)
    SP-->>User: redirect to /

    Note over IdP: next accounts fetch no longer<br/>lists this client_id in approved_clients —<br/>next sign-in needs fresh consent
```

## 4.4 IdP-side self-service revoke (not a FedCM API — plain HTML)

```mermaid
sequenceDiagram
    actor User
    participant IdP as IdP account page (/)

    User->>IdP: visits http://localhost:8080/
    IdP-->>User: profile + "Connected apps" list (from grants table)
    User->>IdP: clicks "Revoke" next to an app
    IdP->>IdP: POST /revoke {client_id} → delete grants row
    IdP-->>User: redirect back to /, app no longer listed
```

This is the mirror image of §4.3: the same `grants` row can be deleted either by the RP calling
`IdentityCredential.disconnect()`, or by the user acting directly on the IdP's own account page.
The second path matters when the user doesn't want to visit every RP individually to revoke
access.
