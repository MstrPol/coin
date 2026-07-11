export type Role = "admin" | "publisher" | "reader";

const rank: Record<Role, number> = {
  reader: 1,
  publisher: 2,
  admin: 3,
};

export function hasMinRole(roles: Role[], min: Role): boolean {
  const need = rank[min];
  return roles.some((r) => rank[r] >= need);
}

export function highestRole(roles: Role[]): Role {
  if (roles.includes("admin")) return "admin";
  if (roles.includes("publisher")) return "publisher";
  return "reader";
}
