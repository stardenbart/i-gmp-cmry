import { cn } from "@/lib/utils"
import { Activity, Wrench, Warehouse, User, Calendar, Clock } from "lucide-react"
import { Card } from "../ui/card"

interface InfoCardProps {
  issue: {
    issue_pic_user_id: string
    pic_name?: string
    due_date: string | null
    created_at: string
    issue_status: string
    computed_status?: string
    follow_up_delay?: number
    habit_name?: string
    equipment_name?: string
    infrastructure_name?: string
    hei?: {
      habit?: { habit_name: string; habit_code?: string }
      equipment?: { equipment_name: string; equipment_code?: string }
      infrastructure?: { infrastructure_name: string; infrastructure_code?: string }
    }
  }
  dueDate?: Date | null
}

export const InfoCard = ({ issue, dueDate }: InfoCardProps) => {
  const isOverdue =
    issue.computed_status === "OpenOverdue" ||
    issue.computed_status === "ClosedOverdue" ||
    (dueDate && dueDate < new Date() && issue.issue_status !== "Closed" && issue.issue_status !== "Verified");

  const habitName = issue.hei?.habit?.habit_name || issue.habit_name;
  const equipmentName = issue.hei?.equipment?.equipment_name || issue.equipment_name;
  const infrastructureName = issue.hei?.infrastructure?.infrastructure_name || issue.infrastructure_name;

  return (
    <Card className="p-6 bg-card/60 backdrop-blur-md lg:col-span-1 space-y-4 h-fit border-border/80 shadow-sm">
      <h3 className="font-semibold border-b border-border pb-2 text-base">Informasi Temuan</h3>

      <div className="space-y-3.5 text-sm">
        {habitName && (
          <div className="flex items-start gap-3">
            <Activity className="h-4 w-4 text-emerald-500 mt-0.5 shrink-0" />
            <div>
              <span className="text-xs text-muted-foreground block font-medium">Klasifikasi Habit</span>
              <span className="font-semibold text-foreground">{habitName}</span>
            </div>
          </div>
        )}

        {equipmentName && (
          <div className="flex items-start gap-3">
            <Wrench className="h-4 w-4 text-blue-500 mt-0.5 shrink-0" />
            <div>
              <span className="text-xs text-muted-foreground block font-medium">Klasifikasi Equipment</span>
              <span className="font-semibold text-foreground">{equipmentName}</span>
            </div>
          </div>
        )}

        {infrastructureName && (
          <div className="flex items-start gap-3">
            <Warehouse className="h-4 w-4 text-purple-500 mt-0.5 shrink-0" />
            <div>
              <span className="text-xs text-muted-foreground block font-medium">Klasifikasi Infrastructure</span>
              <span className="font-semibold text-foreground">{infrastructureName}</span>
            </div>
          </div>
        )}

        <div className="flex items-start gap-3">
          <User className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0" />
          <div>
            <span className="text-xs text-muted-foreground block font-medium">PIC (Penanggung Jawab)</span>
            <span className="font-semibold text-foreground">{issue.pic_name || issue.issue_pic_user_id}</span>
          </div>
        </div>

        <div className="flex items-start gap-3">
          <Calendar className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0" />
          <div>
            <span className="text-xs text-muted-foreground block font-medium">Target Penyelesaian</span>
            {dueDate ? (
              <span className={cn("font-semibold", isOverdue ? "text-red-500 font-bold" : "text-foreground")}>
                {dueDate.toLocaleDateString("id-ID", { day: "numeric", month: "long", year: "numeric" })}
              </span>
            ) : (
              <span className="text-muted-foreground italic">Tidak ditentukan</span>
            )}
          </div>
        </div>

        {issue.follow_up_delay !== undefined && issue.follow_up_delay > 0 && (
          <div className="flex items-start gap-3">
            <Clock className="h-4 w-4 text-orange-500 mt-0.5 shrink-0" />
            <div>
              <span className="text-xs text-muted-foreground block font-medium">Keterlambatan Audit</span>
              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-bold bg-orange-500/10 text-orange-600 border border-orange-500/20">
                Terlambat {issue.follow_up_delay} hari
              </span>
            </div>
          </div>
        )}

        <div className="flex items-start gap-3">
          <Clock className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0" />
          <div>
            <span className="text-xs text-muted-foreground block font-medium">Dibuat Pada</span>
            <span className="font-semibold text-foreground">
              {new Date(issue.created_at).toLocaleDateString("id-ID", { day: "numeric", month: "long", year: "numeric" })}
            </span>
          </div>
        </div>
      </div>
    </Card>
  );
};