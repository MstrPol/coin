import { NavLink, Outlet } from "react-router-dom";
import { highestRole } from "../lib/roles";
import { useAuth } from "../context/AuthContext";

const linkClass = ({ isActive }: { isActive: boolean }) =>
  isActive ? "text-sky-400" : "text-zinc-400 hover:text-zinc-200";

export default function Layout() {
  const { logout, roles, subject, can } = useAuth();
  const role = highestRole(roles);

  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b border-zinc-800 bg-zinc-900/80 backdrop-blur">
        <div className="mx-auto flex max-w-5xl items-center gap-6 px-6 py-4">
          <span className="text-lg font-semibold tracking-tight">Coin</span>
          <nav className="flex gap-4 text-sm">
            <NavLink to="/" className={linkClass} end>
              Dashboard
            </NavLink>
            <NavLink to="/projects" className={linkClass}>
              Projects
            </NavLink>
            <NavLink to="/releases" className={linkClass}>
              GP Releases
            </NavLink>
            <NavLink to="/catalog" className={linkClass}>
              Catalog
            </NavLink>
            <NavLink to="/resolve" className={linkClass}>
              Resolve
            </NavLink>
            <NavLink to="/canary" className={linkClass}>
              Canary
            </NavLink>
            <NavLink to="/components" className={linkClass}>
              Components
            </NavLink>
            {can("publisher") && (
              <NavLink to="/releases/publish" className={linkClass}>
                Publish
              </NavLink>
            )}
            <NavLink to="/audit" className={linkClass}>
              Audit
            </NavLink>
          </nav>
          <div className="ml-auto flex items-center gap-3">
            {roles.length > 0 && (
              <span className="text-xs text-zinc-500">
                {subject ?? "user"} · <span className="font-mono text-zinc-400">{role}</span>
              </span>
            )}
            <button
              type="button"
              onClick={logout}
              className="text-xs text-zinc-500 hover:text-zinc-300"
            >
              Выйти
            </button>
          </div>
        </div>
      </header>
      <main className="mx-auto w-full max-w-5xl flex-1 px-6 py-8">
        <Outlet />
      </main>
    </div>
  );
}
