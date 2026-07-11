import type {
  CatalogOverview,
  CompositionItem,
  HealthSummary,
  Project,
} from "../api/types";
import { pilotProjects } from "./promoteGate";

export type CanaryComponentPin = {
  type: string;
  name: string;
  version: string;
};

export type PromoteCheck = {
  id: string;
  label: string;
  ok: boolean;
  detail: string;
};

export type PromoteWizardPlan = {
  gpName: string;
  catalog: CatalogOverview;
  currentLatest: string;
  canaryVersion: string;
  composition: CompositionItem[];
  canaryComponents: CanaryComponentPin[];
  health: HealthSummary | null;
  pilots: Project[];
  checks: PromoteCheck[];
  ready: boolean;
  blockers: string[];
};

export function buildPromoteChecks(input: {
  canaryVersion: string;
  currentLatest: string;
  health: HealthSummary | null;
  pilots: Project[];
  canaryComponents: CanaryComponentPin[];
}): { checks: PromoteCheck[]; ready: boolean; blockers: string[] } {
  const blockers: string[] = [];
  const checks: PromoteCheck[] = [];

  const hasCanaryLine = !!input.canaryVersion;
  checks.push({
    id: "canary-line",
    label: "Canary line (catalog.latest_canary)",
    ok: hasCanaryLine,
    detail: hasCanaryLine ? input.canaryVersion : "latest_canary не задан",
  });
  if (!hasCanaryLine) blockers.push("Задайте latest_canary в GP Policy");

  const healthOk = input.health != null && input.health.health !== "critical";
  checks.push({
    id: "health",
    label: "Health gate",
    ok: healthOk,
    detail: input.health
      ? `${input.health.health} — ${input.health.successCount} ok / ${input.health.failureCount} fail`
      : "нет build reports за 24h на canary line",
  });
  if (!healthOk) {
    blockers.push(
      input.health?.health === "critical"
        ? "Health critical — исправьте canary builds"
        : "Недостаточно health-данных по canary line",
    );
  }

  const pilotsOk = input.pilots.length > 0;
  checks.push({
    id: "pilots",
    label: "Pilot projects",
    ok: pilotsOk,
    detail: pilotsOk
      ? input.pilots.map((p) => p.name).join(", ")
      : "назначьте canary_mode=canary на Projects или в Studio",
  });
  if (!pilotsOk) blockers.push("Нужен ≥1 pilot project");

  const catalogPromote = hasCanaryLine && input.currentLatest !== input.canaryVersion;
  if (catalogPromote) {
    checks.push({
      id: "catalog",
      label: "Catalog: latest_canary → latest",
      ok: true,
      detail: `${input.currentLatest || "—"} → ${input.canaryVersion}`,
    });
  }

  if (input.canaryComponents.length > 0) {
    checks.push({
      id: "components",
      label: `Components: canary → published (${input.canaryComponents.length})`,
      ok: true,
      detail: input.canaryComponents
        .map((c) => `${c.type}/${c.name}@${c.version}`)
        .join(", "),
    });
  }

  const hasWork = catalogPromote || input.canaryComponents.length > 0;
  if (hasCanaryLine && !hasWork) {
    blockers.push("Stable line уже совпадает с canary и нет canary-компонентов");
  }

  const gateOk = checks.filter((c) => c.id !== "catalog" && c.id !== "components").every((c) => c.ok);
  const ready = hasCanaryLine && hasWork && gateOk;

  return { checks, ready, blockers };
}

export function summarizePlan(plan: PromoteWizardPlan): string[] {
  const steps: string[] = [];
  for (const c of plan.canaryComponents) {
    steps.push(`Promote component ${c.type}/${c.name}@${c.version} → published`);
  }
  if (plan.currentLatest !== plan.canaryVersion) {
    steps.push(`Catalog ${plan.gpName}: latest ← ${plan.canaryVersion}`);
  }
  return steps;
}

export { pilotProjects };
