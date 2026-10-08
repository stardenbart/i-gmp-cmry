import Image from "next/image";

// Shared look for the screens outside the app shell (login, forgot
// password): moving brand gradient, one frosted 400px card, centred content.
export const AUTH_LABEL = "mb-1.5 block text-[11px] font-semibold uppercase tracking-wide text-muted-foreground";
export const AUTH_INPUT = "h-11 rounded-lg border-[1.5px] px-3.5 text-sm";
export const AUTH_ERROR = "mt-1 text-xs font-medium text-destructive";
export const AUTH_BUTTON = "login-btn h-11 w-full rounded-lg text-sm font-semibold text-white shadow-none hover:bg-transparent";

export function AuthCard({
  title,
  subtitle,
  icon,
  children,
  hint,
}: {
  title: string;
  subtitle: React.ReactNode;
  icon?: React.ReactNode;
  children: React.ReactNode;
  hint?: React.ReactNode;
}) {
  return (
    <div className="login-page flex min-h-dvh items-center justify-center px-4 py-10">
      <div className="login-card w-full max-w-[400px] px-6 py-8 text-center text-card-foreground sm:px-9 sm:py-10">
        <div className="mb-4 flex justify-center">
          {icon ?? (
            <Image
              src="/Logo_Cimory.png"
              alt="Cimory"
              width={250}
              height={133}
              priority
              className="h-auto w-[200px] object-contain sm:w-[250px]"
            />
          )}
        </div>
        <h1 className="text-[22px] font-bold text-primary">{title}</h1>
        <p className="mb-6 mt-1 text-[12.5px] text-muted-foreground">{subtitle}</p>

        <div className="text-left">{children}</div>

        {hint && <div className="mt-4 text-center text-xs text-muted-foreground">{hint}</div>}
        <p className="mt-6 text-[11px] text-muted-foreground">Powered by Digital Transformation Plant Sentul</p>
      </div>
    </div>
  );
}
