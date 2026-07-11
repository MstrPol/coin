import { UserManager, WebStorageStateStore } from "oidc-client-ts";

export const oidcEnabled = Boolean(
  import.meta.env.VITE_OIDC_AUTHORITY && import.meta.env.VITE_OIDC_CLIENT_ID,
);

let manager: UserManager | null = null;

export function getUserManager(): UserManager {
  if (!oidcEnabled) {
    throw new Error("OIDC not configured");
  }
  if (!manager) {
    manager = new UserManager({
      authority: import.meta.env.VITE_OIDC_AUTHORITY as string,
      client_id: import.meta.env.VITE_OIDC_CLIENT_ID as string,
      redirect_uri: `${window.location.origin}/login/callback`,
      response_type: "code",
      scope: (import.meta.env.VITE_OIDC_SCOPE as string) || "openid profile email roles",
      userStore: new WebStorageStateStore({ store: window.localStorage }),
    });
  }
  return manager;
}

export async function signInWithOidc(): Promise<void> {
  await getUserManager().signinRedirect();
}

export async function completeOidcCallback() {
  const user = await getUserManager().signinRedirectCallback();
  return user;
}

export async function signOutOidc(): Promise<void> {
  if (!oidcEnabled) return;
  try {
    await getUserManager().signoutRedirect();
  } catch {
    await getUserManager().removeUser();
  }
}
