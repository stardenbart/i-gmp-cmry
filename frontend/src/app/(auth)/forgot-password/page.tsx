"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { ShieldCheck, ArrowLeft } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api/axios";
import { getApiErrorMessage } from "@/lib/api/error";

const forgotPasswordSchema = z.object({
  email: z.string().email("Format email tidak valid"),
});

type ForgotPasswordFormValues = z.infer<typeof forgotPasswordSchema>;

export default function ForgotPasswordPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ForgotPasswordFormValues>({
    resolver: zodResolver(forgotPasswordSchema),
  });

  const onSubmit = async (data: ForgotPasswordFormValues) => {
    try {
      setIsLoading(true);
      const res = await api.post("/auth/forgot-password", data);
      
      setIsSuccess(true);
      toast.success(res.data?.message || "Instruksi reset password telah dikirim");
    } catch (error) {
      toast.error(
        getApiErrorMessage(error, "Gagal mengirim permintaan reset password.")
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-[#050505] px-4 py-12 sm:px-6 lg:px-8">
      {/* Ambient Background Elements */}
      <div className="pointer-events-none absolute -left-[10%] -top-[10%] h-[50vw] max-h-[600px] w-[50vw] max-w-[600px] rounded-full bg-primary/20 blur-[100px] lg:blur-[140px]" />
      <div className="pointer-events-none absolute -right-[10%] bottom-[0%] h-[40vw] max-h-[500px] w-[40vw] max-w-[500px] rounded-full bg-purple-600/20 blur-[100px] lg:blur-[140px]" />
      <div className="pointer-events-none absolute left-[20%] top-[40%] h-[30vw] max-h-[400px] w-[30vw] max-w-[400px] rounded-full bg-emerald-500/10 blur-[120px]" />

      <div className="relative z-10 w-full max-w-[440px]">
        {/* Header Section */}
        <div className="mb-10 text-center">
          <div className="mx-auto mb-6 flex h-20 w-20 items-center justify-center rounded-3xl border border-white/20 bg-gradient-to-tr from-primary/80 to-purple-600/80 shadow-[0_0_40px_rgba(99,102,241,0.5)] backdrop-blur-xl">
            <ShieldCheck className="h-10 w-10 text-white" />
          </div>
          <h2 className="text-3xl font-extrabold tracking-tight text-white sm:text-4xl drop-shadow-sm">
            Lupa Password
          </h2>
          <p className="mt-3 text-sm text-zinc-400 sm:text-base">
            Masukkan email yang terdaftar untuk mengatur ulang password Anda
          </p>
        </div>

        {/* Form Card */}
        <div className="rounded-[2rem] border border-white/10 bg-white/[0.04] p-6 shadow-2xl backdrop-blur-2xl sm:p-10">
          {!isSuccess ? (
            <form className="space-y-6" onSubmit={handleSubmit(onSubmit)}>
              
              {/* Email Field */}
              <div className="space-y-2.5">
                <label
                  htmlFor="email"
                  className="ml-1 text-xs font-semibold uppercase tracking-wider text-zinc-400"
                >
                  Alamat Email
                </label>
                <Input
                  id="email"
                  type="email"
                  autoComplete="email"
                  className="h-12 rounded-2xl border-white/10 bg-black/40 px-4 text-white placeholder:text-zinc-600 focus-visible:border-primary/50 focus-visible:bg-black/60 focus-visible:ring-1 focus-visible:ring-primary/50 transition-all"
                  placeholder="admin@perusahaan.com"
                  {...register("email")}
                />
                {errors.email && (
                  <p className="ml-1 mt-1 text-xs font-medium text-red-400">
                    {errors.email.message}
                  </p>
                )}
              </div>

              {/* Submit Button */}
              <div className="pt-2">
                <Button
                  type="submit"
                  className="h-12 w-full rounded-2xl bg-gradient-to-r from-primary to-purple-600 text-[15px] font-bold text-white shadow-lg shadow-primary/25 transition-all hover:scale-[1.02] hover:from-primary/90 hover:to-purple-600/90 active:scale-[0.98]"
                  isLoading={isLoading}
                >
                  Kirim Instruksi Reset
                </Button>
              </div>

            </form>
          ) : (
            <div className="text-center space-y-4">
              <div className="mx-auto w-16 h-16 bg-green-500/20 text-green-500 rounded-full flex items-center justify-center mb-4">
                <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7"></path></svg>
              </div>
              <h3 className="text-xl font-bold text-white">Cek Email Anda</h3>
              <p className="text-zinc-400 text-sm">
                Kami telah mengirimkan instruksi dan tautan untuk mengatur ulang password ke alamat email Anda.
              </p>
            </div>
          )}
          
          <div className="mt-8 text-center">
            <button 
              onClick={() => router.push("/login")}
              className="inline-flex items-center text-sm font-medium text-zinc-400 hover:text-white transition-colors"
            >
              <ArrowLeft className="w-4 h-4 mr-2" />
              Kembali ke Halaman Login
            </button>
          </div>
        </div>
        
        {/* Footer Text */}
        <p className="mt-8 text-center text-xs text-zinc-500">
          Dilindungi oleh sistem enkripsi standar industri.
        </p>
      </div>
    </div>
  );
}
