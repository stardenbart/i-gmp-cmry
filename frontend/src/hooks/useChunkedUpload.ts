import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { issueApi } from "@/lib/api/issue.api";
import { convertToWebP } from "@/lib/utils/imageUtils";

export interface ChunkedUploadProps {
  issueId: string;
  onSuccessCallback?: () => void;
  onErrorCallback?: () => void;
}

export const useChunkedUpload = ({ issueId, onSuccessCallback, onErrorCallback }: ChunkedUploadProps) => {
  const queryClient = useQueryClient();
  const [uploadProgress, setUploadProgress] = useState<number>(0);

  const mutation = useMutation({
    mutationFn: async ({
      file: originalFile,
      type,
      refPhotoId,
    }: {
      file: File;
      type: "Initial" | "FollowUp" | "WOWR";
      refPhotoId?: string;
    }) => {
      // 1. Convert image to WebP (if it is an image)
      const file = await convertToWebP(originalFile);

      const CHUNK_SIZE = 2 * 1024 * 1024; // 2MB
      const totalChunks = Math.ceil(file.size / CHUNK_SIZE);
      const fileId = Math.random().toString(36).substring(2, 15) + Date.now().toString(36);

      let lastResponse = null;

      for (let i = 0; i < totalChunks; i++) {
        const start = i * CHUNK_SIZE;
        const end = Math.min(start + CHUNK_SIZE, file.size);
        const chunk = file.slice(start, end);

        const formData = new FormData();
        formData.append("photo", new File([chunk], file.name, { type: file.type }));
        formData.append("photo_type", type);
        if (refPhotoId) {
          formData.append("ref_photo_id", refPhotoId);
        }
        formData.append("chunk_index", i.toString());
        formData.append("total_chunks", totalChunks.toString());
        formData.append("file_id", fileId);

        lastResponse = await issueApi.uploadPhoto(issueId, formData);
        
        // Update progress
        const progress = Math.round(((i + 1) / totalChunks) * 100);
        setUploadProgress(progress);
      }
      return lastResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issue-photos", issueId] });
      toast.success("Foto berhasil diunggah");
      setUploadProgress(0);
      if (onSuccessCallback) onSuccessCallback();
    },
    onError: () => {
      toast.error("Gagal mengunggah foto");
      setUploadProgress(0);
      if (onErrorCallback) onErrorCallback();
    },
  });

  return {
    uploadMutation: mutation,
    uploadProgress,
  };
};
