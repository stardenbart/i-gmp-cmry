import { RoleBadge } from "../RoleBadge";

const USE_CARDS = [
  { n: "01", title: "Susuri lewat bab", desc: "Chip bab di atas mengikuti urutan menu sungguhan di aplikasi." },
  { n: "02", title: "Saring per peran", desc: "Tekan salah satu tombol peran untuk menyembunyikan bagian yang bukan urusan Anda." },
  { n: "03", title: "Loncat cepat", desc: "Klik chip bab mana pun untuk langsung melompat ke bagian itu." },
];

export function WelcomeSection() {
  return (
    <section id="selamat-datang" className="scroll-mt-32 space-y-6">
      <div className="space-y-3">
        <p className="font-mono text-xs uppercase tracking-wider text-primary">Panduan Pengguna</p>
        <h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">I-GMP</h1>
        <p className="max-w-2xl text-base text-muted-foreground">
          Panduan ini menjelaskan cara memakai seluruh bagian aplikasi audit &amp; inspeksi — dari mengisi
          checklist inspeksi, menindaklanjuti temuan, mengajukan perintah kerja, sampai mengatur data induk
          dan pengguna. Disusun mengikuti urutan menu yang sama dengan aplikasi.
        </p>
      </div>

      <p className="max-w-2xl text-sm text-muted-foreground">
        Karena satu aplikasi ini dipakai tiga peran berbeda —{" "}
        <strong className="font-semibold text-foreground">Admin/Super Admin</strong>,{" "}
        <strong className="font-semibold text-foreground">Auditor</strong>, dan{" "}
        <strong className="font-semibold text-foreground">Auditee/PIC</strong> (Penanggung Jawab) — setiap
        bagian diberi label peran. Gunakan tombol saring di bagian atas halaman untuk hanya menampilkan
        bagian yang relevan dengan peran Anda.
      </p>

      <div className="flex flex-wrap gap-2">
        <RoleBadge role="all" />
        <RoleBadge role="admin" />
        <RoleBadge role="auditor" />
        <RoleBadge role="auditee" />
      </div>

      <div className="grid gap-3 sm:grid-cols-3">
        {USE_CARDS.map((c) => (
          <div key={c.n} className="rounded-3xl border border-border bg-card p-4">
            <p className="font-mono text-xs text-primary">{c.n}</p>
            <h3 className="mt-1 text-sm font-semibold text-foreground">{c.title}</h3>
            <p className="mt-1 text-xs text-muted-foreground">{c.desc}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
