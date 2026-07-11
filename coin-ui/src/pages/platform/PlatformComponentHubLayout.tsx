import { Link, NavLink, Outlet, useParams } from "react-router-dom";
import { useAuth } from "../../context/AuthContext";
import {
  familyCatalogPath,
  familyNewDraftPath,
  PLATFORM_FAMILIES,
  type PlatformFamilyId,
} from "../../lib/platformFamilyConfig";

const tabs = [
  { label: "Overview", segment: "" },
  { label: "Releases", segment: "releases" },
] as const;

function tabClass({ isActive }: { isActive: boolean }) {
  return `border-b-2 px-4 py-2 text-sm whitespace-nowrap ${
    isActive
      ? "border-sky-500 text-sky-400"
      : "border-transparent text-zinc-500 hover:text-zinc-300"
  }`;
}

export default function PlatformComponentHubLayout({ familyId }: { familyId: PlatformFamilyId }) {
  const { name = "" } = useParams();
  const { can } = useAuth();
  const family = PLATFORM_FAMILIES[familyId];
  const base = `${familyCatalogPath(familyId)}/${encodeURIComponent(name)}`;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <Link to={familyCatalogPath(familyId)} className="text-sm text-sky-400 hover:underline">
            ← {family.catalogTitle}
          </Link>
          <h1 className="mt-2 text-2xl font-semibold font-mono">{name}</h1>
          <p className="mt-1 text-zinc-400">{family.profileLabel}</p>
        </div>
        {can("publisher") && (
          <Link
            to={familyNewDraftPath(familyId, name)}
            className="rounded-lg bg-sky-600 px-4 py-2 text-sm font-semibold text-white hover:bg-sky-500"
          >
            New draft
          </Link>
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

      <Outlet context={{ familyId, compType: family.compType }} />
    </div>
  );
}
