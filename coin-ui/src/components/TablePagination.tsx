const PAGE_SIZES = [25, 50, 100] as const;

type Props = {
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
};

export function pageToOffset(page: number, pageSize: number): number {
  return Math.max(0, (page - 1) * pageSize);
}

export function totalPages(total: number, pageSize: number): number {
  if (pageSize <= 0) return 1;
  return Math.max(1, Math.ceil(total / pageSize));
}

export default function TablePagination({
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
}: Props) {
  const pages = totalPages(total, pageSize);
  const from = total === 0 ? 0 : pageToOffset(page, pageSize) + 1;
  const to = Math.min(page * pageSize, total);

  return (
    <div className="flex flex-wrap items-center justify-between gap-4 text-sm text-zinc-400">
      <span>
        {total === 0 ? "0 записей" : `${from}–${to} из ${total}`}
      </span>
      <div className="flex flex-wrap items-center gap-3">
        <label className="flex items-center gap-2">
          <span className="text-xs">На странице</span>
          <select
            value={pageSize}
            onChange={(e) => onPageSizeChange(Number(e.target.value))}
            className="rounded border border-zinc-700 bg-zinc-950 px-2 py-1 text-sm"
          >
            {PAGE_SIZES.map((n) => (
              <option key={n} value={n}>
                {n}
              </option>
            ))}
          </select>
        </label>
        <span>
          Страница {page} из {pages}
        </span>
        <div className="flex gap-1">
          <button
            type="button"
            disabled={page <= 1}
            onClick={() => onPageChange(page - 1)}
            className="rounded border border-zinc-700 px-3 py-1 hover:bg-zinc-800 disabled:opacity-40"
          >
            ←
          </button>
          <button
            type="button"
            disabled={page >= pages}
            onClick={() => onPageChange(page + 1)}
            className="rounded border border-zinc-700 px-3 py-1 hover:bg-zinc-800 disabled:opacity-40"
          >
            →
          </button>
        </div>
      </div>
    </div>
  );
}
