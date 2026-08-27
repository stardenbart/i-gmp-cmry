export default function DashboardLoading() {
  return (
    <div className="animate-pulse space-y-6" role="status" aria-label="Memuat halaman">
      <div className="space-y-2">
        <div className="h-7 w-52 rounded-lg bg-muted" />
        <div className="h-4 w-72 max-w-full rounded bg-muted/70" />
      </div>
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, index) => (
          <div key={index} className="h-28 rounded-2xl border border-border bg-card" />
        ))}
      </div>
      <div className="h-72 rounded-2xl border border-border bg-card" />
      <span className="sr-only">Memuat konten…</span>
    </div>
  );
}
