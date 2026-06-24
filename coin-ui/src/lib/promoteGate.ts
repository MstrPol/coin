import type { HealthSummary, Project } from "../api/types";

export type GpHealthRow = {
  gpName: string;
  gpVersion: string;
  health: HealthSummary | null;
};

export function pilotProjects(projects: Project[]): Project[] {
  return projects.filter((p) => p.canaryMode === "canary");
}

export function promoteGate(input: {
  healthRows: GpHealthRow[];
  projects: Project[];
  selectedPilotNames: Set<string>;
}): { ok: boolean; reasons: string[] } {
  const reasons: string[] = [];
  const pilots = pilotProjects(input.projects);
  const pendingPilots = input.selectedPilotNames.size;
  if (pilots.length === 0 && pendingPilots === 0) {
    reasons.push("Выберите хотя бы один pilot project (canary_mode=canary)");
  }
  if (input.healthRows.length === 0) {
    reasons.push("Компонент не привязан к GP release — проверьте composition");
  }
  for (const row of input.healthRows) {
    if (!row.health) {
      reasons.push(`Нет health-данных для ${row.gpName}@${row.gpVersion}`);
      continue;
    }
    if (row.health.health === "critical") {
      reasons.push(`Health critical для ${row.gpName}@${row.gpVersion} — promote заблокирован`);
    }
  }
  const hasHealthySignal = input.healthRows.some((r) => r.health && r.health.health !== "critical");
  if (input.healthRows.length > 0 && !hasHealthySignal) {
    reasons.push("Нет GP canary line с приемлемым health");
  }
  return { ok: reasons.length === 0, reasons };
}

export function healthBadgeClass(health: HealthSummary["health"]): string {
  switch (health) {
    case "healthy":
      return "bg-emerald-950/50 text-emerald-400";
    case "degraded":
      return "bg-amber-950/50 text-amber-400";
    case "critical":
      return "bg-red-950/50 text-red-400";
    default:
      return "bg-zinc-800 text-zinc-400";
  }
}
