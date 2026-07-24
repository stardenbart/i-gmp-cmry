import { cn } from "@/lib/utils"
import { FileText, User, Calendar, Clock } from "lucide-react"
import { Card } from "../ui/card"

interface InfoCardProps {
  issue: {
    keterangan: string | null
    issue_pic_user_id: string
    pic_name?: string
    due_date: string | null
    created_at: string
    issue_status: string
  },
  dueDate?: Date | null

}

export const InfoCard = ({ issue, dueDate }: InfoCardProps) => {
    return (
        <Card className="p-6 bg-card/60 backdrop-blur-md lg:col-span-1 space-y-4 h-fit">
          <h3 className="font-semibold border-b border-border pb-2">Informasi Temuan</h3>
          
          <div className="space-y-3 text-sm">
            <div className="flex items-start gap-3">
              <FileText className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0" />
              <div>
                <span className="text-xs text-muted-foreground block">Keterangan</span>
                <span className="font-medium">{issue.keterangan || "-"}</span>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <User className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0" />
              <div>
                <span className="text-xs text-muted-foreground block">PIC (Penanggung Jawab)</span>
                <span className="font-medium">{issue.pic_name || issue.issue_pic_user_id}</span>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <Calendar className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0" />
              <div>
                <span className="text-xs text-muted-foreground block">Target Penyelesaian</span>
                {dueDate ? (
                  <span className={cn("font-medium", dueDate < new Date() && issue.issue_status !== "Closed" && issue.issue_status !== "Verified" ? "text-red-500" : "")}>
                    {dueDate.toLocaleDateString("id-ID", { day: "numeric", month: "long", year: "numeric" })}
                  </span>
                ) : (
                  <span className="text-muted-foreground">Tidak ditentukan</span>
                )}
              </div>
            </div>
            <div className="flex items-start gap-3">
              <Clock className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0" />
              <div>
                <span className="text-xs text-muted-foreground block">Dibuat</span>
                <span className="font-medium">
                  {new Date(issue.created_at).toLocaleDateString("id-ID", { day: "numeric", month: "long", year: "numeric" })}
                </span>
              </div>
            </div>
          </div>
        </Card>
    )
}