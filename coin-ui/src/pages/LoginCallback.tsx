import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { completeOidcCallback } from "../lib/oidc";

export default function LoginCallback() {
  const { loginWithToken } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    completeOidcCallback()
      .then(async (user) => {
        const token = user.access_token;
        if (!token) {
          throw new Error("missing access_token");
        }
        await loginWithToken(token);
        navigate("/", { replace: true });
      })
      .catch((err: Error) => setError(err.message));
  }, [loginWithToken, navigate]);

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center p-8">
        <p className="text-red-400">SSO error: {error}</p>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-8">
      <p className="text-zinc-400">Завершение входа SSO…</p>
    </div>
  );
}
