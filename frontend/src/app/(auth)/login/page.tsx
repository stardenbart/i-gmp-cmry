"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { AUTH_BUTTON, AUTH_ERROR, AUTH_INPUT, AUTH_LABEL, AuthCard } from "@/components/auth/AuthCard";
import { api } from "@/lib/api/axios";
import { getApiErrorMessage } from "@/lib/api/error";
import { useAuthStore } from "@/stores/authStore";

const loginSchema = z.object({
  username: z.string().trim().min(3, "Username atau email minimal 3 karakter"),
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
    if (isLoading) return;

    setIsLoading(true);

    try {
      const res = await api.post("/auth/login", data);

      // The access/refresh tokens never reach this code — the backend sets
      // them as httpOnly cookies on the login response. Only the user
      // profile comes back in the body.
      const backendUser = res.data.data.user;
      const user = {
        ...backendUser,
        id: backendUser.user_id,
        name: backendUser.full_name,
      };

      const plantCode = backendUser.plant_id || "global";
      setAuth(user);
      
      toast.success("Login berhasil!");
      router.replace(`/cimory/${plantCode}/dashboard/${user.id}`);
    } catch (error) {
      setIsLoading(false);
      toast.error(
        getApiErrorMessage(error, "Login gagal. Periksa kembali username/email dan password Anda.")
      );
    }
  };

  return (
    <AuthCard
      title="I-GMP"
      subtitle="PT Cisarua Mountain Dairy, Tbk — Plant Sentul"
      hint="Masuk untuk mengakses dasbor inspeksi Anda."
    >
      <form className="space-y-4" onSubmit={handleSubmit(onSubmit)}>
        <div>
          <label htmlFor="username" className={AUTH_LABEL}>
            Username atau Email
          </label>
          <Input
            id="username"
            type="text"
            autoComplete="username"
            aria-required="true"
            aria-invalid={errors.username ? "true" : "false"}
            aria-describedby={errors.username ? "username-error" : undefined}
            className={AUTH_INPUT}
            placeholder="admin atau nama@perusahaan.com"
            {...register("username")}
          />
          {errors.username && (
            <p id="username-error" className={AUTH_ERROR}>
              {errors.username.message}
            </p>
          )}
        </div>

        <div>
          <div className="flex items-center justify-between">
            <label htmlFor="password" className={AUTH_LABEL}>
              Password
            </label>
            <Link
              href="/forgot-password"
              className="mb-1.5 rounded-sm text-xs font-semibold text-info hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              aria-label="Lupa password akun Anda"
            >
              Lupa password?
            </Link>
          </div>
          <Input
            id="password"
            type="password"
            autoComplete="current-password"
            aria-required="true"
            aria-invalid={errors.password ? "true" : "false"}
            aria-describedby={errors.password ? "password-error" : undefined}
            className={AUTH_INPUT}
            placeholder="••••••••"
            {...register("password")}
          />
          {errors.password && (
            <p id="password-error" className={AUTH_ERROR}>
              {errors.password.message}
            </p>
          )}
        </div>

        <Button type="submit" className={AUTH_BUTTON} isLoading={isLoading}>
          Masuk
        </Button>
      </form>
    </AuthCard>
  );
}
