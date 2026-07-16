"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { ShieldCheck } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api/axios";
import { useAuthStore } from "@/stores/authStore";

const loginSchema = z.object({
  email: z.string().email("Format email tidak valid"),
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
      const user = res.data.data.user;
      
      setAuth(token, user);
      
      toast.success("Login berhasil!");
      router.push(`/cimory/dashboard/${user.id}`);
    } catch (error: any) {
      toast.error(
        error.response?.data?.message || "Login gagal. Periksa kembali email dan password Anda."
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-black px-4 sm:px-6 lg:px-8">
      {/* Abstract Background Elements (Stitch inspired) */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-[20%] -left-[10%] w-[50%] h-[50%] rounded-full bg-primary/20 blur-[120px]" />
        <div className="absolute bottom-[10%] -right-[10%] w-[40%] h-[40%] rounded-full bg-purple-500/20 blur-[120px]" />
      </div>

      <div className="w-full max-w-md space-y-8 z-10">
        <div className="text-center">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-3xl bg-card border border-border shadow-2xl">
            <ShieldCheck className="h-8 w-8 text-primary" />
          </div>
          <h2 className="mt-6 text-3xl font-bold tracking-tight text-foreground">
            Audit Monitoring
          </h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Masuk untuk mengakses dasbor inspeksi
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit(onSubmit)}>
          <div className="space-y-4 rounded-3xl bg-card/50 backdrop-blur-xl border border-border p-8 shadow-2xl">
            <div>
              <label
                htmlFor="email"
                className="block text-xs font-medium text-muted-foreground ml-1 mb-2"
              >
                Email address
              </label>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                placeholder="auditor@company.com"
                {...register("email")}
              />
              {errors.email && (
                <p className="mt-1 text-xs text-red-500 ml-1">
                  {errors.email.message}
                </p>
              )}
            </div>

            <div>
              <div className="flex items-center justify-between ml-1 mb-2">
                <label
                  htmlFor="password"
                  className="block text-xs font-medium text-muted-foreground"
                >
                  Password
                </label>
                <div className="text-xs">
                  <a
                    href="/forgot-password"
                    className="font-medium text-primary hover:text-primary/80"
                  >
                    Lupa password?
                  </a>
                </div>
              </div>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                placeholder="••••••••"
                {...register("password")}
              />
              {errors.password && (
                <p className="mt-1 text-xs text-red-500 ml-1">
                  {errors.password.message}
                </p>
              )}
            </div>

            <div className="pt-2">
              <Button type="submit" className="w-full" isLoading={isLoading} size="lg">
                Sign in
              </Button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
