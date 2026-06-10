import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import type { Role } from "../api/types";
import { api, getAccessToken, getApiKey, setAccessToken, setApiKey } from "../lib/api";
import { hasMinRole } from "../lib/roles";
import { isSkipLocal, logoutSession } from "../lib/session";

type AuthContextValue = {
  apiKey: string | null;
  roles: Role[];
  subject: string | null;
  loginWithKey: (key: string) => Promise<void>;
  loginWithToken: (token: string, subject?: string) => Promise<void>;
  logout: () => void;
  can: (min: Role) => boolean;
  refreshMe: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [apiKey, setKeyState] = useState<string | null>(() => getApiKey());
  const [roles, setRoles] = useState<Role[]>([]);
  const [subject, setSubject] = useState<string | null>(null);

  const applyMe = useCallback((me: Awaited<ReturnType<typeof api.me>>) => {
    setRoles(me.roles);
    setSubject(me.subject);
  }, []);

  const refreshMe = useCallback(async () => {
    if (!getApiKey() && !getAccessToken() && !isSkipLocal()) {
      setRoles([]);
      setSubject(null);
      return;
    }
    const me = await api.me();
    applyMe(me);
  }, [applyMe]);

  useEffect(() => {
    if (getApiKey() || getAccessToken() || isSkipLocal()) {
      refreshMe().catch(() => {
        logoutSession();
        setKeyState(null);
        setRoles([]);
        setSubject(null);
      });
    }
  }, [refreshMe]);

  const loginWithKey = useCallback(
    async (key: string) => {
      setApiKey(key.trim());
      setKeyState(key.trim());
      const me = await api.me();
      applyMe(me);
    },
    [applyMe],
  );

  const loginWithToken = useCallback(
    async (token: string) => {
      setAccessToken(token);
      setKeyState(null);
      const me = await api.me();
      applyMe(me);
    },
    [applyMe],
  );

  const logout = useCallback(() => {
    logoutSession();
    setKeyState(null);
    setRoles([]);
    setSubject(null);
  }, []);

  const can = useCallback((min: Role) => hasMinRole(roles, min), [roles]);

  const value = useMemo(
    () => ({
      apiKey,
      roles,
      subject,
      loginWithKey,
      loginWithToken,
      logout,
      can,
      refreshMe,
    }),
    [apiKey, roles, subject, loginWithKey, loginWithToken, logout, can, refreshMe],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth outside AuthProvider");
  return ctx;
}
