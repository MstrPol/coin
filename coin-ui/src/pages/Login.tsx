import { FormEvent, useState } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { setApiKey } from "../lib/api";
import { oidcEnabled, signInWithOidc } from "../lib/oidc";
import { setSkipFlag } from "../lib/session";
import { useAuth } from "../context/AuthContext";
import { isAuthenticated } from "../components/RequireAuth";

export default function Login() {
  const { loginWithKey, refreshMe } = useAuth();
  const navigate = useNavigate();
  const [key, setKey] = useState("dev-local-admin-key");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  if (isAuthenticated()) {
    return <Navigate to="/" replace />;
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    const trimmed = key.trim();
    setApiKey(trimmed);
    try {
      await loginWithKey(trimmed);
      navigate("/");
    } catch (err) {
      localStorage.removeItem("coin-admin-api-key");
      setError(err instanceof Error ? err.message : "auth failed");
    } finally {
      setLoading(false);
    }
  }

  async function skipLocal() {
    setSkipFlag();
    try {
      await refreshMe();
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "local dev auth failed");
    }
  }

  async function onOidc() {
    setError(null);
    try {
      await signInWithOidc();
    } catch (err) {
      setError(err instanceof Error ? err.message : "oidc redirect failed");
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center px-4">
      <form
        onSubmit={onSubmit}
        className="w-full max-w-md space-y-4 rounded-lg border border-zinc-800 bg-zinc-900 p-8"
      >
        <h1 className="text-xl font-semibold">Coin Control Plane</h1>
        <p className="text-sm text-zinc-400">
          Admin API key (<code className="text-zinc-300">X-API-Key</code>) или SSO
        </p>

        {oidcEnabled && (
          <button
            type="button"
            onClick={onOidc}
            className="w-full rounded border border-zinc-600 py-2 text-sm hover:bg-zinc-800"
          >
            Войти через SSO
          </button>
        )}

        <input
          type="password"
          value={key}
          onChange={(e) => setKey(e.target.value)}
          className="w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
          placeholder="COIN_ADMIN_API_KEY"
          autoComplete="off"
        />

        <p className="text-xs text-zinc-500">
          Local RBAC keys: admin / publisher / reader — см.{" "}
          <code className="text-zinc-400">COIN_*_API_KEY</code> в docker/.env
        </p>

        {error && <p className="text-sm text-red-400">{error}</p>}

        <button
          type="submit"
          disabled={loading}
          className="w-full rounded bg-sky-600 py-2 text-sm font-medium hover:bg-sky-500 disabled:opacity-50"
        >
          {loading ? "…" : "Войти по API key"}
        </button>

        <button
          type="button"
          onClick={skipLocal}
          className="w-full text-sm text-zinc-500 hover:text-zinc-300"
        >
          Пропустить (local, AUTH_DISABLED)
        </button>
      </form>
    </div>
  );
}
