import { Navigate, Outlet } from "react-router-dom";
import type { Role } from "../api/types";
import { useAuth } from "../context/AuthContext";
import { isAuthenticated } from "./RequireAuth";

type Props = {
  min: Role;
};

export default function RequireRole({ min }: Props) {
  const { can, roles } = useAuth();

  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />;
  }

  if (roles.length === 0) {
    return <p className="text-zinc-500">Проверка прав…</p>;
  }

  if (!can(min)) {
    return (
      <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-8">
        <h1 className="text-lg font-semibold">Доступ запрещён</h1>
        <p className="mt-2 text-sm text-zinc-400">
          Для этой страницы нужна роль <span className="font-mono text-zinc-200">{min}</span> или
          выше.
        </p>
      </div>
    );
  }

  return <Outlet />;
}
