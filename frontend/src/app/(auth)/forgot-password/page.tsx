"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { ArrowLeft, CheckCircle2, KeyRound, Mail } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { AUTH_BUTTON, AUTH_ERROR, AUTH_INPUT, AUTH_LABEL, AuthCard } from "@/components/auth/AuthCard";
import { cn } from "@/lib/utils";
import { api } from "@/lib/api/axios";
import { getApiErrorMessage, getApiErrorStatus } from "@/lib/api/error";

const requestOTPSchema = z.object({
  email: z.string().trim().email("Format email tidak valid"),
});

const resetPasswordSchema = z.object({
  otp: z.string().regex(/^\d{6}$/, "OTP harus terdiri dari 6 angka"),
  newPassword: z.string().min(8, "Password minimal 8 karakter"),
  confirmPassword: z.string().min(8, "Konfirmasi password minimal 8 karakter"),
}).refine((values) => values.newPassword === values.confirmPassword, {
  message: "Konfirmasi password tidak sama",
  path: ["confirmPassword"],
});

type RequestOTPValues = z.infer<typeof requestOTPSchema>;
type ResetPasswordValues = z.infer<typeof resetPasswordSchema>;
type ResetStep = "request" | "verify" | "success";

export default function ForgotPasswordPage() {
  const router = useRouter();
  const [step, setStep] = useState<ResetStep>("request");
  const [email, setEmail] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isResending, setIsResending] = useState(false);

  const requestForm = useForm<RequestOTPValues>({ resolver: zodResolver(requestOTPSchema) });
  const resetForm = useForm<ResetPasswordValues>({
    resolver: zodResolver(resetPasswordSchema),
    defaultValues: { otp: "", newPassword: "", confirmPassword: "" },
  });

  const requestOTP = async (targetEmail: string) => {
    const response = await api.post("/auth/forgot-password", { email: targetEmail });
    return response.data?.message as string | undefined;
  };

  const onRequestOTP = async ({ email: requestedEmail }: RequestOTPValues) => {
    try {
      setIsLoading(true);
      const normalizedEmail = requestedEmail.trim();
      const message = await requestOTP(normalizedEmail);
      setEmail(normalizedEmail);
      setStep("verify");
      toast.success(message || "Jika email terdaftar, kode OTP telah dikirim");
    } catch (error) {
      const message = getApiErrorMessage(error, "Gagal mengirim kode OTP.");
      if (getApiErrorStatus(error) === 404) {
        requestForm.setError("email", { type: "server", message: "Email tidak terdaftar" });
      }
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  const onResendOTP = async () => {
    try {
      setIsResending(true);
      const message = await requestOTP(email);
      resetForm.setValue("otp", "");
      toast.success(message || "Kode OTP baru telah dikirim");
    } catch (error) {
      toast.error(getApiErrorMessage(error, "Gagal mengirim ulang kode OTP."));
    } finally {
      setIsResending(false);
    }
  };

  const onResetPassword = async (values: ResetPasswordValues) => {
    try {
      setIsLoading(true);
      const response = await api.post("/auth/reset-password", {
        email,
        otp: values.otp,
        new_password: values.newPassword,
        confirm_password: values.confirmPassword,
      });
      setStep("success");
      toast.success(response.data?.message || "Password berhasil diubah");
    } catch (error) {
      toast.error(getApiErrorMessage(error, "OTP tidak valid atau sudah kedaluwarsa."));
    } finally {
      setIsLoading(false);
    }
  };

  const title = step === "request" ? "Lupa Password" : step === "verify" ? "Verifikasi OTP" : "Password Diubah";
  const subtitle =
    step === "request"
      ? "Masukkan email terdaftar untuk menerima kode OTP"
      : step === "verify"
        ? "Masukkan OTP dari email lalu buat password baru"
        : "Password baru Anda sudah dapat digunakan";

  return (
    <AuthCard
      title={title}
      subtitle={subtitle}
      icon={step === "success" ? <CheckCircle2 aria-hidden="true" className="h-14 w-14 text-success" /> : undefined}
      hint="OTP disimpan sebagai hash satu arah dan hanya dapat digunakan sekali."
    >
      {step === "request" && (
        <form className="space-y-4" onSubmit={requestForm.handleSubmit(onRequestOTP)}>
          <div>
            <label htmlFor="email" className={AUTH_LABEL}>Alamat Email</label>
            <div className="relative">
              <Mail aria-hidden="true" className="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input id="email" type="email" autoComplete="email" className={cn(AUTH_INPUT, "pl-10")} placeholder="nama@perusahaan.com" {...requestForm.register("email")} />
            </div>
            {requestForm.formState.errors.email && <p className={AUTH_ERROR}>{requestForm.formState.errors.email.message}</p>}
          </div>
          <Button type="submit" className={AUTH_BUTTON} isLoading={isLoading}>Kirim Kode OTP</Button>
        </form>
      )}

      {step === "verify" && (
        <form className="space-y-4" onSubmit={resetForm.handleSubmit(onResetPassword)}>
          <div className="rounded-sm border-l-4 border-info bg-info/10 px-3.5 py-2.5 text-[13px] leading-relaxed text-foreground">
            OTP berlaku 10 menit dan dikirim ke <strong>{email}</strong>.
          </div>

          <div>
            <label htmlFor="otp" className={AUTH_LABEL}>Kode OTP</label>
            <Input id="otp" inputMode="numeric" autoComplete="one-time-code" maxLength={6} className={cn(AUTH_INPUT, "h-12 text-center font-mono text-xl tracking-[0.45em] placeholder:tracking-normal")} placeholder="000000" {...resetForm.register("otp", { onChange: (event) => { event.target.value = event.target.value.replace(/\D/g, "").slice(0, 6); } })} />
            {resetForm.formState.errors.otp && <p className={AUTH_ERROR}>{resetForm.formState.errors.otp.message}</p>}
          </div>

          <div>
            <label htmlFor="new-password" className={AUTH_LABEL}>Password Baru</label>
            <Input id="new-password" type="password" autoComplete="new-password" className={AUTH_INPUT} {...resetForm.register("newPassword")} />
            {resetForm.formState.errors.newPassword && <p className={AUTH_ERROR}>{resetForm.formState.errors.newPassword.message}</p>}
          </div>

          <div>
            <label htmlFor="confirm-password" className={AUTH_LABEL}>Konfirmasi Password</label>
            <Input id="confirm-password" type="password" autoComplete="new-password" className={AUTH_INPUT} {...resetForm.register("confirmPassword")} />
            {resetForm.formState.errors.confirmPassword && <p className={AUTH_ERROR}>{resetForm.formState.errors.confirmPassword.message}</p>}
          </div>

          <Button type="submit" className={AUTH_BUTTON} isLoading={isLoading}>
            <KeyRound aria-hidden="true" className="h-4 w-4" /> Atur Password Baru
          </Button>

          <div className="flex flex-wrap items-center justify-between gap-3 text-xs">
            <button type="button" onClick={() => setStep("request")} className="text-muted-foreground hover:text-foreground">Ganti alamat email</button>
            <button type="button" onClick={onResendOTP} disabled={isResending} className="rounded-sm px-2 py-1 font-semibold text-info hover:underline disabled:opacity-50">{isResending ? "Mengirim..." : "Kirim ulang OTP"}</button>
          </div>
        </form>
      )}

      {step === "success" && (
        <div className="space-y-4 text-center">
          <p className="text-sm text-muted-foreground">Silakan masuk menggunakan password baru Anda.</p>
          <Button type="button" onClick={() => router.replace("/login")} className={AUTH_BUTTON}>Masuk Sekarang</Button>
        </div>
      )}

      {step !== "success" && (
        <div className="mt-6 text-center">
          <button type="button" onClick={() => router.push("/login")} className="inline-flex items-center text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"><ArrowLeft aria-hidden="true" className="mr-2 h-4 w-4" />Kembali ke Halaman Login</button>
        </div>
      )}
    </AuthCard>
  );
}
