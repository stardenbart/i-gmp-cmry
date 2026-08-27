"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { ArrowLeft, CheckCircle2, KeyRound, Mail, ShieldCheck } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-[#050505] px-4 py-12 sm:px-6 lg:px-8">
      <div className="pointer-events-none absolute -left-[10%] -top-[10%] h-[50vw] max-h-[600px] w-[50vw] max-w-[600px] rounded-full bg-primary/20 blur-[100px] lg:blur-[140px]" />
      <div className="pointer-events-none absolute -right-[10%] bottom-0 h-[40vw] max-h-[500px] w-[40vw] max-w-[500px] rounded-full bg-purple-600/20 blur-[100px] lg:blur-[140px]" />
      <div className="pointer-events-none absolute left-[20%] top-[40%] h-[30vw] max-h-[400px] w-[30vw] max-w-[400px] rounded-full bg-emerald-500/10 blur-[120px]" />

      <div className="relative z-10 w-full max-w-[440px]">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-5 flex h-20 w-20 items-center justify-center rounded-3xl border border-white/20 bg-gradient-to-tr from-primary/80 to-purple-600/80 shadow-[0_0_40px_rgba(99,102,241,0.5)] backdrop-blur-xl">
            {step === "success" ? <CheckCircle2 className="h-10 w-10 text-white" /> : <ShieldCheck className="h-10 w-10 text-white" />}
          </div>
          <h1 className="text-3xl font-extrabold tracking-tight text-white sm:text-4xl">
            {step === "request" ? "Lupa Password" : step === "verify" ? "Verifikasi OTP" : "Password Diubah"}
          </h1>
          <p className="mt-3 text-sm text-zinc-400 sm:text-base">
            {step === "request" && "Masukkan email terdaftar untuk menerima kode OTP"}
            {step === "verify" && "Masukkan OTP dari email lalu buat password baru"}
            {step === "success" && "Password baru Anda sudah dapat digunakan"}
          </p>
        </div>

        <div className="rounded-[2rem] border border-white/10 bg-white/[0.04] p-6 shadow-2xl backdrop-blur-2xl sm:p-10">
          {step === "request" && (
            <form className="space-y-6" onSubmit={requestForm.handleSubmit(onRequestOTP)}>
              <div className="space-y-2.5">
                <label htmlFor="email" className="ml-1 text-xs font-semibold uppercase tracking-wider text-zinc-400">Alamat Email</label>
                <div className="relative">
                  <Mail aria-hidden="true" className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-zinc-500" />
                  <Input id="email" type="email" autoComplete="email" className="h-12 border-white/10 bg-black/40 pl-11 text-white placeholder:text-zinc-600 focus-visible:border-primary/50 focus-visible:bg-black/60" placeholder="nama@perusahaan.com" {...requestForm.register("email")} />
                </div>
                {requestForm.formState.errors.email && <p className="ml-1 text-xs font-medium text-red-400">{requestForm.formState.errors.email.message}</p>}
              </div>
              <Button type="submit" className="h-12 w-full rounded-2xl bg-gradient-to-r from-primary to-purple-600 font-bold text-white" isLoading={isLoading}>Kirim Kode OTP</Button>
            </form>
          )}

          {step === "verify" && (
            <form className="space-y-5" onSubmit={resetForm.handleSubmit(onResetPassword)}>
              <div className="rounded-xl border border-primary/20 bg-primary/10 px-4 py-3 text-xs text-zinc-300">
                OTP berlaku 10 menit dan dikirim ke <strong className="text-white">{email}</strong>.
              </div>

              <div className="space-y-2">
                <label htmlFor="otp" className="ml-1 text-xs font-semibold uppercase tracking-wider text-zinc-400">Kode OTP</label>
                <Input id="otp" inputMode="numeric" autoComplete="one-time-code" maxLength={6} className="h-14 border-white/10 bg-black/40 text-center font-mono text-xl tracking-[0.45em] text-white placeholder:tracking-normal" placeholder="000000" {...resetForm.register("otp", { onChange: (event) => { event.target.value = event.target.value.replace(/\D/g, "").slice(0, 6); } })} />
                {resetForm.formState.errors.otp && <p className="text-xs font-medium text-red-400">{resetForm.formState.errors.otp.message}</p>}
              </div>

              <div className="space-y-2">
                <label htmlFor="new-password" className="ml-1 text-xs font-semibold uppercase tracking-wider text-zinc-400">Password Baru</label>
                <Input id="new-password" type="password" autoComplete="new-password" className="border-white/10 bg-black/40 text-white" {...resetForm.register("newPassword")} />
                {resetForm.formState.errors.newPassword && <p className="text-xs font-medium text-red-400">{resetForm.formState.errors.newPassword.message}</p>}
              </div>

              <div className="space-y-2">
                <label htmlFor="confirm-password" className="ml-1 text-xs font-semibold uppercase tracking-wider text-zinc-400">Konfirmasi Password</label>
                <Input id="confirm-password" type="password" autoComplete="new-password" className="border-white/10 bg-black/40 text-white" {...resetForm.register("confirmPassword")} />
                {resetForm.formState.errors.confirmPassword && <p className="text-xs font-medium text-red-400">{resetForm.formState.errors.confirmPassword.message}</p>}
              </div>

              <Button type="submit" className="h-12 w-full rounded-2xl bg-gradient-to-r from-primary to-purple-600 font-bold text-white" isLoading={isLoading}>
                <KeyRound aria-hidden="true" className="mr-2 h-4 w-4" /> Atur Password Baru
              </Button>

              <div className="flex flex-wrap items-center justify-between gap-3 text-xs">
                <button type="button" onClick={() => setStep("request")} className="text-zinc-400 hover:text-white">Ganti alamat email</button>
                <button type="button" onClick={onResendOTP} disabled={isResending} className="font-semibold text-primary hover:text-primary/80 disabled:opacity-50">{isResending ? "Mengirim..." : "Kirim ulang OTP"}</button>
              </div>
            </form>
          )}

          {step === "success" && (
            <div className="space-y-5 text-center">
              <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-green-500/20 text-green-400"><CheckCircle2 className="h-8 w-8" /></div>
              <div><h2 className="text-xl font-bold text-white">Password Berhasil Diubah</h2><p className="mt-2 text-sm text-zinc-400">Silakan masuk menggunakan password baru Anda.</p></div>
              <Button type="button" onClick={() => router.replace("/login")} className="h-12 w-full rounded-2xl">Masuk Sekarang</Button>
            </div>
          )}

          {step !== "success" && (
            <div className="mt-8 text-center">
              <button type="button" onClick={() => router.push("/login")} className="inline-flex items-center text-sm font-medium text-zinc-400 transition-colors hover:text-white"><ArrowLeft aria-hidden="true" className="mr-2 h-4 w-4" />Kembali ke Halaman Login</button>
            </div>
          )}
        </div>

        <p className="mt-8 text-center text-xs text-zinc-500">OTP disimpan sebagai hash satu arah dan hanya dapat digunakan sekali.</p>
      </div>
    </div>
  );
}
