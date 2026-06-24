import { Link, NavLink, Outlet, useParams } from "react-router-dom";
import { useAuth } from "../../context/AuthContext";

const tabs = [
  { label: "Overview", segment: "" },
  { label: "Releases", segment: "releases" },
  { label: "Policy", segment: "policy" },
  { label: "Canary", segment: "canary" },
  { label: "Build stack", segment: "build-stack" },
] as const;

function tabClass({ isActive }: { isActive: boolean }) {
  return `border-b-2 px-4 py-2 text-sm whitespace-nowrap ${
    isActive
      ? "border-sky-500 text-sky-400"
      : "border-transparent text-zinc-500 hover:text-zinc-300"
  }`;
}

export default function GpHubLayout() {
  const { name = "" } = useParams();
  const { can } = useAuth();
  const base = `/gp/${encodeURIComponent(name)}`;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <Link to="/gp" className="text-sm text-sky-400 hover:underline">
            ← GP Profiles
          </Link>
          <h1 className="mt-2 text-2xl font-semibold font-mono">{name}</h1>
          <p className="mt-1 text-zinc-400">Golden Path profile</p>
        </div>
        {can("publisher") && (
          <div className="flex shrink-0 flex-wrap gap-2">
            <Link
              to={`${base}/releases/new-draft`}
              className="rounded-lg border border-zinc-700 px-4 py-2 text-sm hover:bg-zinc-800"
            >
              New draft
            </Link>
            <Link
              to={`${base}/releases/new`}
              className="rounded-lg bg-sky-600 px-4 py-2 text-sm font-semibold text-white hover:bg-sky-500"
            >
              New release
            </Link>
          </div>
        )}
      </div>

      <nav className="flex gap-1 overflow-x-auto border-b border-zinc-800">
        {tabs.map((t) => (
          <NavLink
            key={t.segment || "overview"}
            to={t.segment ? `${base}/${t.segment}` : base}
            end={t.segment === ""}
            className={tabClass}
          >
            {t.label}
          </NavLink>
        ))}
      </nav>

      <Outlet />
    </div>
  );
}
