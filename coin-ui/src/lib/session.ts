import { clearAuthCredentials } from "./api";

export const SKIP_KEY = "coin-admin-api-key-skipped";

export function isSkipLocal(): boolean {
  return localStorage.getItem(SKIP_KEY) === "1";
}

export function setSkipFlag(): void {
  localStorage.setItem(SKIP_KEY, "1");
}

export function clearSkipFlag(): void {
  localStorage.removeItem(SKIP_KEY);
}

export function logoutSession(): void {
  clearAuthCredentials();
  clearSkipFlag();
}
