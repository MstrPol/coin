import { useCallback, useEffect, useState } from "react";
import GpContentEditor from "./studio/GpContentEditor";
import type { GpContentModel } from "../lib/gpContentYaml";
import { emptyGpContentModel, pipelineBodyToModel } from "../lib/gpContentYaml";
import { api } from "../lib/api";

type Props = {
  gpName: string;
  version: string;
  canEdit: boolean;
};

export default function GpReleasePipelineEditor({ gpName, version, canEdit }: Props) {
  const [model, setModel] = useState<GpContentModel | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    api
      .getGPReleasePipeline(gpName, version)
      .then((body) => setModel(pipelineBodyToModel(body, gpName)))
      .catch(() => setModel(emptyGpContentModel(gpName)))
      .finally(() => setLoading(false));
  }, [gpName, version]);

  const previewState = useGpReleasePipelinePreview(model, gpName, version, !canEdit);

  async function savePipeline() {
    if (!model || !canEdit) return;
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      await api.saveGPReleasePipeline(gpName, version, model);
      setMessage("Pipeline сохранён");
    } catch (err) {
      setError(err instanceof Error ? err.message : "save failed");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return <p className="text-sm text-zinc-500">Загрузка pipeline…</p>;
  }

  if (!model) {
    return <p className="text-sm text-red-400">Не удалось загрузить pipeline</p>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-4">
        <h2 className="text-lg font-medium">Pipeline</h2>
        {canEdit && (
          <button
            type="button"
            onClick={() => void savePipeline()}
            disabled={saving}
            className="rounded bg-zinc-700 px-3 py-1.5 text-sm hover:bg-zinc-600 disabled:opacity-50"
          >
            {saving ? "Saving…" : "Save pipeline"}
          </button>
        )}
      </div>
      {error && <p className="text-sm text-red-400">{error}</p>}
      {message && <p className="text-sm text-emerald-400">{message}</p>}
      <div className="grid gap-6 lg:grid-cols-[1fr_20rem]">
        <GpContentEditor model={model} onChange={setModel} disabled={!canEdit} />
        <PipelinePreviewPanel {...previewState} />
      </div>
    </div>
  );
}

function useGpReleasePipelinePreview(
  model: GpContentModel | null,
  gpName: string,
  version: string,
  disabled?: boolean,
) {
  const [preview, setPreview] = useState<import("../lib/gpContentYaml").GpContentPreviewResult | null>(
    null,
  );
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [previewing, setPreviewing] = useState(false);

  const runPreview = useCallback(async () => {
    if (!model) {
      setPreview(null);
      setPreviewError(null);
      setPreviewing(false);
      return;
    }
    setPreviewing(true);
    setPreviewError(null);
    try {
      const result = await api.gpReleasePipelinePreview(gpName, version, model);
      setPreview(result);
    } catch (err) {
      setPreviewError(err instanceof Error ? err.message : "preview failed");
      setPreview(null);
    } finally {
      setPreviewing(false);
    }
  }, [model, gpName, version]);

  useEffect(() => {
    if (disabled || !model) return;
    const timer = window.setTimeout(() => void runPreview(), 400);
    return () => window.clearTimeout(timer);
  }, [model, disabled, runPreview]);

  return { preview, previewing, previewError };
}

function PipelinePreviewPanel({
  preview,
  previewing,
  previewError,
}: {
  preview: import("../lib/gpContentYaml").GpContentPreviewResult | null;
  previewing: boolean;
  previewError: string | null;
}) {
  return (
    <aside className="space-y-3 rounded border border-zinc-800 p-4 text-xs">
      <h3 className="font-medium text-zinc-300">Preview</h3>
      {previewing && <p className="text-zinc-500">Обновление…</p>}
      {previewError && <p className="text-red-400">{previewError}</p>}
      {preview && !preview.valid && (
        <ul className="space-y-1 text-amber-300">
          {preview.issues.map((i) => (
            <li key={`${i.field}:${i.message}`}>
              <span className="font-mono text-zinc-500">{i.field}</span> {i.message}
            </li>
          ))}
        </ul>
      )}
      {preview?.pipeline && (
        <pre className="max-h-[32rem] overflow-auto rounded bg-zinc-950 p-2 font-mono text-[10px] text-zinc-400">
          {JSON.stringify(preview.pipeline, null, 2)}
        </pre>
      )}
    </aside>
  );
}
