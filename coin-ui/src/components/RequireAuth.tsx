import { Navigate, Outlet, useLocation } from "react-router-dom";
import { getAccessToken, getApiKey } from "../lib/api";
import { isSkipLocal } from "../lib/session";

export function isAuthenticated(): boolean {
  return Boolean(getApiKey() || getAccessToken()) || isSkipLocal();
}

export default function RequireAuth() {
  const location = useLocation();
  if (!isAuthenticated()) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  return <Outlet />;
}

export { setSkipFlag, clearSkipFlag, isSkipLocal } from "../lib/session";
