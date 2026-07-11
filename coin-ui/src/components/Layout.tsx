import { NavLink, Outlet } from "react-router-dom";
import { highestRole } from "../lib/roles";
import { visibleNavGroups } from "../lib/nav";
import { useAuth } from "../context/AuthContext";

const linkClass = ({ isActive }: { isActive: boolean }) =>
  isActive
    ? "bg-zinc-800 text-sky-400"
    : "text-zinc-400 hover:bg-zinc-800/60 hover:text-zinc-200";

export default function Layout() {
  const { logout, roles, subject } = useAuth();
  const role = highestRole(roles);
  const groups = visibleNavGroups(roles);

  return (
    <div className="min-h-screen flex">
      <aside className="flex w-60 shrink-0 flex-col border-r border-zinc-800 bg-zinc-900/50">
        <div className="border-b border-zinc-800 px-4 py-5">
          <span className="text-lg font-semibold tracking-tight">Coin</span>
          <p className="mt-0.5 text-xs text-zinc-500">Enabling console</p>
        </div>

        <nav className="flex-1 overflow-y-auto px-2 py-4 space-y-6">
          {groups.map((group) => (
            <div key={group.id}>
              <p className="mb-2 px-2 text-[10px] font-semibold uppercase tracking-wider text-zinc-600">
                {group.label}
              </p>
              <ul className="space-y-0.5">
                {group.items.map((item) => (
                  <li key={item.to}>
                    <NavLink
                      to={item.to}
                      end={item.end}
                      className={({ isActive }) =>
                        `block rounded-md px-2 py-1.5 text-sm ${linkClass({ isActive })}`
                      }
                    >
                      {item.label}
                    </NavLink>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </nav>

      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex items-center gap-4 border-b border-zinc-800 bg-zinc-900/40 px-6 py-3">
          <div className="ml-auto flex items-center gap-3">
            <a
              href="/api/docs/"
              target="_blank"
              rel="noopener noreferrer"
              className="text-xs text-zinc-500 hover:text-zinc-300"
            >
              API docs ↗
            </a>
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
        </header>
        <main className="flex-1 overflow-x-auto px-6 py-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
