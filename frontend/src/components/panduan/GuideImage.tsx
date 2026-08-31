import Image from "next/image";

export function GuideImage({
  src,
  alt,
  caption,
  width = 1440,
  height = 900,
}: {
  src: string;
  alt: string;
  caption?: string;
  width?: number;
  height?: number;
}) {
  return (
    <figure className="overflow-hidden rounded-3xl border border-border bg-card shadow-sm">
      <Image
        src={src}
        alt={alt}
        width={width}
        height={height}
        className="h-auto w-full object-cover object-top"
        sizes="(min-width: 1024px) 800px, 100vw"
      />
      {caption && (
        <figcaption className="border-t border-border px-4 py-2.5 text-xs text-muted-foreground">{caption}</figcaption>
      )}
    </figure>
  );
}
