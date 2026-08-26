"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import Image from "next/image";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api/axios";
import { getApiErrorMessage } from "@/lib/api/error";
import { useAuthStore } from "@/stores/authStore";

const loginSchema = z.object({
  username: z.string().min(3, "Username minimal 3 karakter"),
  password: z.string().min(6, "Password minimal 6 karakter"),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const router = useRouter();
  const setAuth = useAuthStore((state) => state.setAuth);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginFormValues) => {
    try {
      setIsLoading(true);
      const res = await api.post("/auth/login", data);
      
      const token = res.data.data.token;
      const backendUser = res.data.data.user;
      const user = {
        ...backendUser,
        id: backendUser.user_id,
        name: backendUser.full_name,
      };
      
      const plantCode = backendUser.plant_id || "global";
      setAuth(token, user);
      
      toast.success("Login berhasil!");
      router.push(`/cimory/${plantCode}/dashboard/${user.id}`);
    } catch (error) {
      toast.error(
        getApiErrorMessage(error, "Login gagal. Periksa kembali username dan password Anda.")
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-[#050505] px-4 py-12 sm:px-6 lg:px-8">
      {/* Lightweight Ambient Background Elements */}
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(59,130,246,0.15),transparent_50%)]" />
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_bottom_right,rgba(147,51,234,0.15),transparent_50%)]" />

      <div className="relative z-10 w-full max-w-110">
        {/* Header Section */}
        <div className="mb-10 text-center flex flex-col items-center">
          <div className="mb-4 flex items-center justify-center">
            <Image 
              src="/Logo_Cimory.png" 
              alt="Cimory Logo" 
              width={200} 
              height={80} 
              priority
              className="h-12 sm:h-14 w-auto object-contain" 
            />
          </div>
          <p className="mt-6 text-sm text-zinc-400 sm:text-base">
            Masuk untuk mengakses dasbor inspeksi Anda
          </p>
        </div>

        {/* Login Form Card */}
        <div className="rounded-[2rem] border border-white/10 bg-white/[0.04] p-6 shadow-2xl backdrop-blur-2xl sm:p-10">
          <form className="space-y-6" onSubmit={handleSubmit(onSubmit)}>
            
            {/* Username Field */}
            <div className="space-y-2.5">
              <label
                htmlFor="username"
                className="ml-1 text-xs font-semibold uppercase tracking-wider text-zinc-300"
              >
                Username
              </label>
              <Input
                id="username"
                type="text"
                autoComplete="username"
                aria-required="true"
                aria-invalid={errors.username ? "true" : "false"}
                aria-describedby={errors.username ? "username-error" : undefined}
                className="h-12 rounded-2xl border-white/10 bg-black/40 px-4 text-white placeholder:text-zinc-500 focus-visible:border-primary/50 focus-visible:bg-black/60 focus-visible:ring-1 focus-visible:ring-primary/50 transition-all"
                placeholder="admin"
                {...register("username")}
              />
              {errors.username && (
                <p id="username-error" className="ml-1 mt-1 text-xs font-medium text-red-400">
                  {errors.username.message}
                </p>
              )}
            </div>

            {/* Password Field */}
            <div className="space-y-2.5">
              <div className="ml-1 flex items-center justify-between">
                <label
                  htmlFor="password"
                  className="text-xs font-semibold uppercase tracking-wider text-zinc-300"
                >
                  Password
                </label>
                <a
                  href="/forgot-password"
                  className="text-xs font-medium text-primary hover:text-primary/80 transition-colors focus:outline-none focus:ring-1 focus:ring-primary"
                  aria-label="Lupa password akun Anda"
                >
                  Lupa password?
                </a>
              </div>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                aria-required="true"
                aria-invalid={errors.password ? "true" : "false"}
                aria-describedby={errors.password ? "password-error" : undefined}
                className="h-12 rounded-2xl border-white/10 bg-black/40 px-4 text-white placeholder:text-zinc-500 focus-visible:border-primary/50 focus-visible:bg-black/60 focus-visible:ring-1 focus-visible:ring-primary/50 transition-all"
                placeholder="••••••••"
                {...register("password")}
              />
              {errors.password && (
                <p id="password-error" className="ml-1 mt-1 text-xs font-medium text-red-400">
                  {errors.password.message}
                </p>
              )}
            </div>

            {/* Submit Button */}
            <div className="pt-2">
              <Button
                type="submit"
                aria-label="Sign In ke Akun"
                className="h-12 w-full min-h-[48px] rounded-2xl bg-gradient-to-r from-primary to-purple-600 text-[15px] font-bold text-white shadow-lg shadow-primary/25 transition-all hover:scale-[1.02] hover:from-primary/90 hover:to-purple-600/90 active:scale-[0.98]"
                isLoading={isLoading}
              >
                Sign In
              </Button>
            </div>

          </form>
        </div>
        
        {/* Footer Text */}
        <p className="mt-8 text-center text-xs text-zinc-500">
          Dilindungi oleh sistem enkripsi standar industri.
        </p>
      </div>
    </div>
  );
}
