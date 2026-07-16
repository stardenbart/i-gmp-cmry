import Link from "next/link";
import { Card } from "../ui/card";
import { ClipboardCheck, AlertTriangle, Users, History } from "lucide-react";

const ACTIONS = [
  {
    href: "inspections/create",
    icon: ClipboardCheck,
    title: "Inspeksi Baru",
    description: "Buat inspeksi",
    iconBg: "bg-primary/10",
    iconColor: "text-primary",
  },
  {
    href: "issues",
    icon: AlertTriangle,
    title: "Lihat Issues",
    description: "Semua temuan",
    iconBg: "bg-orange-500/10",
    iconColor: "text-orange-500",
  },
  {
    href: "master",
    icon: Users,
    title: "Master Data",
    description: "Kelola data",
    iconBg: "bg-purple-500/10",
    iconColor: "text-purple-500",
  },
  {
    href: "logs",
    icon: History,
    title: "Audit Trail",
    description: "Riwayat",
    iconBg: "bg-green-500/10",
    iconColor: "text-green-500",
  },
];

export const QuickActions = () => {
  return (
    <Card className="p-6 bg-card/60 backdrop-blur-md">
      <div className="mb-4">
        <h3 className="font-semibold">Quick Actions</h3>
        <p className="text-xs text-muted-foreground">Aksi cepat</p>
      </div>
      <div className="grid grid-cols-2 gap-3">
        {ACTIONS.map((action, idx) => {
          const Icon = action.icon;
          return (
            <Link
              key={idx}
              href={action.href}
              className="flex items-center gap-3 p-3 rounded-lg border border-border hover:bg-muted/50 transition-colors"
            >
              <div className={`p-2 rounded-lg ${action.iconBg}`}>
                <Icon className={`h-4 w-4 ${action.iconColor}`} />
              </div>
              <div>
                <span className="text-sm font-medium block">{action.title}</span>
                <span className="text-xs text-muted-foreground">{action.description}</span>
              </div>
            </Link>
          );
        })}
      </div>
    </Card>
  );
};