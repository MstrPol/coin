import type { Role } from "./roles";
import { hasMinRole } from "./roles";

export type NavItem = {
  label: string;
  to: string;
  end?: boolean;
  minRole?: Role;
};

export type NavGroup = {
  id: string;
  label: string;
  items: NavItem[];
  minRole?: Role;
};

export const navGroups: NavGroup[] = [
  {
    id: "overview",
    label: "Overview",
    items: [{ label: "Dashboard", to: "/", end: true }],
  },
  {
    id: "fleet",
    label: "Fleet",
    items: [
      { label: "Projects", to: "/projects" },
      { label: "Build reports", to: "/build-reports" },
    ],
  },
  {
    id: "golden-paths",
    label: "Golden Paths",
    items: [
      { label: "GP Profiles", to: "/gp" },
      { label: "Resolve", to: "/resolve" },
    ],
  },
  {
    id: "platform",
    label: "Platform",
    items: [
      { label: "Runtime", to: "/platform/runtime" },
      { label: "Branching models", to: "/platform/branching-models" },
    ],
  },
  {
    id: "admin",
    label: "Admin",
    items: [
      { label: "Audit", to: "/audit", minRole: "admin" },
    ],
  },
];

export function visibleNavGroups(roles: Role[]): NavGroup[] {
  return navGroups
    .filter((g) => !g.minRole || hasMinRole(roles, g.minRole))
    .map((g) => ({
      ...g,
      items: g.items.filter((item) => !item.minRole || hasMinRole(roles, item.minRole)),
    }))
    .filter((g) => g.items.length > 0);
}
