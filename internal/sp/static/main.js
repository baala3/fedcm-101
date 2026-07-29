// FedCM client logic for the demo Relying Party. This is the only place in
// the SP that talks FedCM directly — everything else is plain cookie auth.

function setStatus(text, kind) {
  const status = document.getElementById("status");
  if (!status) return;
  status.textContent = text;
  status.className = kind ? "status-" + kind : "";
}

function providerConfig() {
  return {
    configURL: window.FEDCM_IDP_CONFIG_URL,
    clientId: window.FEDCM_CLIENT_ID,
    params: { nonce: crypto.randomUUID() },
  };
}

// Only one navigator.credentials.get() call may be outstanding at a time.
// The silent auto-reauth attempt on page load and a user-triggered click can
// otherwise race, so we abort whichever request is in flight before
// starting a new one.
let activeAbortController = null;

async function requestCredential(mediation) {
  if (activeAbortController) activeAbortController.abort();
  const controller = new AbortController();
  activeAbortController = controller;
  try {
    const credential = await navigator.credentials.get({
      identity: { providers: [providerConfig()] },
      mediation,
      signal: controller.signal,
    });
    return credential;
  } finally {
    if (activeAbortController === controller) activeAbortController = null;
  }
}

async function completeSignIn(credential) {
  const res = await fetch("/session", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ token: credential.token }),
  });
  if (!res.ok) throw new Error("session exchange failed");
  window.location.href = "/profile";
}

async function signIn() {
  const signInBtn = document.getElementById("signin-btn");
  if (signInBtn) signInBtn.disabled = true;
  setStatus("Waiting for account selection…", "info");
  try {
    const credential = await requestCredential("optional");
    await completeSignIn(credential);
  } catch (err) {
    if (err.name === "AbortError") return; // superseded by another request
    console.error("FedCM sign-in failed:", err);
    setStatus("Sign-in failed or was dismissed: " + err.message, "error");
  } finally {
    if (signInBtn) signInBtn.disabled = false;
  }
}

// Called on the home page: try a silent, no-UI sign-in first. This is what
// lets a returning user (who already granted consent) skip the account
// chooser entirely. Chrome/Edge also enforce a quiet period after any
// sign-in (silent or manual) during which silent attempts are always
// declined as an anti-abuse measure — that's expected, not a bug, and just
// falls back to the button below.
async function attemptSilentSignIn() {
  if (!window.IdentityCredential) return;
  setStatus("Checking for an existing sign-in…", "info");
  try {
    const credential = await requestCredential("silent");
    await completeSignIn(credential);
  } catch (err) {
    if (err.name === "AbortError") return; // superseded by another request
    console.log("Silent auto-reauth not available:", err.message);
    setStatus("", null);
  }
}

async function disconnect() {
  const disconnectBtn = document.getElementById("disconnect-btn");
  if (disconnectBtn) disconnectBtn.disabled = true;
  setStatus("Disconnecting…", "info");
  try {
    await IdentityCredential.disconnect({
      configURL: window.FEDCM_IDP_CONFIG_URL,
      clientId: window.FEDCM_CLIENT_ID,
    });
    await fetch("/disconnected", { method: "POST" });
    window.location.href = "/";
  } catch (err) {
    console.error("Disconnect failed:", err);
    setStatus("Disconnect failed: " + err.message, "error");
    if (disconnectBtn) disconnectBtn.disabled = false;
  }
}

document.addEventListener("DOMContentLoaded", () => {
  const signInBtn = document.getElementById("signin-btn");
  if (signInBtn) signInBtn.addEventListener("click", signIn);

  const disconnectBtn = document.getElementById("disconnect-btn");
  if (disconnectBtn) disconnectBtn.addEventListener("click", disconnect);

  if (document.body.dataset.page === "index") {
    attemptSilentSignIn();
  }
});
