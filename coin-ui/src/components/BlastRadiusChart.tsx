import { Link } from "react-router-dom";
import type { BlastRadius } from "../api/types";

type Props = {
  blast: BlastRadius;
  highlightVersion?: string;
};

export default function BlastRadiusChart({ blast, highlightVersion }: Props) {
  const target = highlightVersion ?? blast.version;
  const max = Math.max(...blast.byVersion.map((v) => v.count), 1);

  return (
    <div className="space-y-4">
      <p className="text-sm text-zinc-400">
        {blast.onThisVersion} из {blast.totalOnGP} projects на{" "}
        <span className="font-mono text-zinc-200">
          {blast.goldenPath}@{target}
        </span>
      </p>

      {blast.byVersion.length === 0 ? (
        <p className="text-sm text-zinc-500">Нет projects на этом GP</p>
      ) : (
        <ul className="space-y-2">
          {blast.byVersion.map((v) => {
            const pct = Math.round((v.count / max) * 100);
            const isTarget = v.version === target;
            return (
              <li key={v.version} className="grid grid-cols-[5rem_1fr_3rem] items-center gap-3 text-sm">
                <span
                  className={`font-mono ${isTarget ? "text-sky-400" : "text-zinc-400"}`}
                >
                  {v.version}
                </span>
                <div className="h-6 overflow-hidden rounded bg-zinc-800">
                  <div
                    className={`h-full rounded ${isTarget ? "bg-sky-600" : "bg-zinc-600"}`}
                    style={{ width: `${pct}%`, minWidth: v.count > 0 ? "0.5rem" : 0 }}
                  />
                </div>
                <span className="tabular-nums text-zinc-300">{v.count}</span>
              </li>
            );
          })}
        </ul>
      )}

      <dl className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <Stat label="На этой версии" value={blast.onThisVersion} />
        <Stat label="На других версиях" value={blast.onOtherVersions} />
        <Stat label="На более старых" value={blast.onOlderVersions} />
        <Stat label="Всего на GP" value={blast.totalOnGP} />
      </dl>

      {blast.onThisVersion > 0 && (
        <Link
          to={`/projects?goldenPath=${encodeURIComponent(blast.goldenPath)}&version=${encodeURIComponent(target)}`}
          className="inline-block text-sm text-sky-400 hover:underline"
        >
          Показать projects →
        </Link>
      )}
    </div>
  );
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div>
      <dt className="text-xs text-zinc-500">{label}</dt>
      <dd className="text-2xl font-semibold tabular-nums">{value}</dd>
    </div>
  );
}
