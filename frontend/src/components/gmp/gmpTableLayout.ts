// Column sizing for the Data Inspeksi (GMP) table. The table takes the width
// its columns need (w-max) and scrolls sideways inside the card, so the long
// text columns keep a readable width instead of being squeezed to fit.
export const GMP_TABLE = {
  scroller: "overflow-x-auto [scrollbar-width:thin] [scrollbar-color:var(--border)_transparent]",
  table: "w-max min-w-full text-sm text-left border-collapse",
  uraian: "min-w-[360px] max-w-[440px] whitespace-normal",
  keterangan: "min-w-[320px] max-w-[420px] whitespace-normal",
  followUpKeterangan: "min-w-[320px] max-w-[420px] whitespace-normal",
} as const;
