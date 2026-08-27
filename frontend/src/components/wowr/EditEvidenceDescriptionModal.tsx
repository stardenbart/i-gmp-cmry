"use client";

import { useState } from "react";
import { createPortal } from "react-dom";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { FileText, X } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { issueApi } from "@/lib/api/issue.api";
import { getApiErrorMessage } from "@/lib/api/error";
import { useMounted } from "@/lib/useMounted";

const MAX_DESCRIPTION_LENGTH = 1000;

interface EditEvidenceDescriptionModalProps {
  photoId: string;
  issueId: string;
  initialDescription?: string;
  onClose: () => void;
}

export function EditEvidenceDescriptionModal({
  photoId,
  issueId,
  initialDescription = "",
  onClose,
}: EditEvidenceDescriptionModalProps) {
  const mounted = useMounted();
  const queryClient = useQueryClient();
  const [description, setDescription] = useState(initialDescription);

  const updateMutation = useMutation({
    mutationFn: () => issueApi.updatePhoto(photoId, description.trim()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      queryClient.invalidateQueries({ queryKey: ["issue-photos", issueId] });
      toast.success("Keterangan bukti penyelesaian berhasil diperbarui");
      onClose();
    },
    onError: (error) => {
      toast.error(getApiErrorMessage(error, "Gagal memperbarui keterangan bukti penyelesaian"));
    },
  });

  if (!mounted || typeof window === "undefined") return null;

  return createPortal(
    <div
      className="fixed inset-0 z-[110] flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      aria-labelledby="edit-evidence-description-title"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget && !updateMutation.isPending) onClose();
      }}
    >
      <div className="w-full max-w-lg space-y-4 rounded-2xl border border-border/60 bg-background p-5 shadow-2xl">
        <div className="flex items-center justify-between border-b border-border pb-3">
          <div className="flex min-w-0 items-center gap-2">
            <FileText className="h-4 w-4 shrink-0 text-primary" />
            <h3 id="edit-evidence-description-title" className="truncate text-base font-bold">
              Edit Keterangan Bukti Penyelesaian
            </h3>
          </div>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={onClose}
            disabled={updateMutation.isPending}
            className="h-8 w-8 shrink-0 rounded-full"
            aria-label="Tutup"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>

        <div className="space-y-1.5">
          <label htmlFor={`wowr-evidence-description-${photoId}`} className="text-xs font-semibold text-foreground">
            Keterangan bukti
          </label>
          <textarea
            id={`wowr-evidence-description-${photoId}`}
            value={description}
            onChange={(event) => setDescription(event.target.value)}
            maxLength={MAX_DESCRIPTION_LENGTH}
            rows={5}
            autoFocus
            disabled={updateMutation.isPending}
            placeholder="Jelaskan pekerjaan atau perbaikan yang telah diselesaikan..."
            className="w-full resize-y rounded-xl border border-border bg-card px-3 py-2.5 text-sm leading-relaxed outline-none transition focus:ring-2 focus:ring-primary/40 disabled:opacity-60"
          />
          <p className="text-right text-[10px] text-muted-foreground">
            {description.length}/{MAX_DESCRIPTION_LENGTH}
          </p>
        </div>

        <div className="flex justify-end gap-2 border-t border-border pt-3">
          <Button type="button" variant="outline" onClick={onClose} disabled={updateMutation.isPending}>
            Batal
          </Button>
          <Button type="button" onClick={() => updateMutation.mutate()} isLoading={updateMutation.isPending}>
            Simpan Keterangan
          </Button>
        </div>
      </div>
    </div>,
    document.body
  );
}
