import { useState, useRef } from "react"
import { Upload, ImageIcon, Eye, Trash2 } from "lucide-react"
import { Button } from "../ui/button"
import { Card } from "../ui/card"

// 1. Interface untuk tipe data foto
interface PhotoItem {
    issue_photo_id: string;
    image_url: string;
}

// 2. Props untuk komponen item foto satuan
interface PhotoCardProps {
    photo: PhotoItem;
    onPreview: () => void;
    onDelete: () => void;
    isDeleting: boolean;
    isAuditor?: boolean;
    isInitialPhoto?: boolean;
    canDelete?: boolean;
}

// Komponen Kecil: Menampilkan item foto satuan
const PhotoCard = ({ photo, onPreview, onDelete, isDeleting, isAuditor = true, isInitialPhoto = false, canDelete = true }: PhotoCardProps) => {
    return (
        <div className="group relative aspect-square overflow-hidden rounded-xl border border-border bg-muted">
            <img 
                src={photo.image_url} 
                alt="Dokumentasi" 
                className="h-full w-full object-cover transition-transform group-hover:scale-105"
            />
            {/* Overlay Menu Saat Hover */}
            <div className="absolute inset-0 flex items-center justify-center gap-2 bg-black/40 opacity-0 transition-opacity group-hover:opacity-100">
                <Button size="icon" variant="outline" className="h-8 w-8" onClick={onPreview}>
                    <Eye className="h-4 w-4" />
                </Button>
                {(!isInitialPhoto || isAuditor) && canDelete && (
                    <Button 
                        size="icon" 
                        variant="destructive" 
                        className="h-8 w-8" 
                        onClick={onDelete}
                        disabled={isDeleting}
                    >
                        <Trash2 className="h-4 w-4" />
                    </Button>
                )}
            </div>
        </div>
    )
}

// 3. Props untuk komponen kontainer utama
interface PhotoSectionProps {
    initialPhotos: PhotoItem[];
    followUpPhotos: PhotoItem[];
    uploadMutation: { isPending: boolean; mutate: (data: { file: File; type: "Initial" | "FollowUp" }) => void };
    deleteMutation: { isPending: boolean; mutate: (id: string) => void };
    setSelectedImage: (url: string) => void;
    uploadProgress?: number;
    isAuditor?: boolean;
    canUploadFollowUp?: boolean;
}

// Komponen Utama: Kontainer Dokumentasi Foto
export const PhotoSection = ({
    initialPhotos = [],
    followUpPhotos = [],
    uploadMutation,
    deleteMutation,
    setSelectedImage,
    uploadProgress = 0,
    isAuditor = true,
    canUploadFollowUp = true
}: PhotoSectionProps) => {
    // State internal untuk tipe foto yang dipilih di dropdown. Jika bukan auditor, selalu 'FollowUp'
    const [photoType, setPhotoType] = useState<"Initial" | "FollowUp">(isAuditor ? "Initial" : "FollowUp")
    const fileInputRef = useRef<HTMLInputElement>(null)

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0]
        if (file) {
            uploadMutation.mutate({ file, type: photoType })
            // Reset input file agar bisa upload file yang sama berturut-turut jika dibutuhkan
            e.target.value = ""
        }
    }

    const isUploadDisabled = uploadMutation.isPending || (!isAuditor && !canUploadFollowUp);

    return (
        <Card className="p-6 bg-card/60 backdrop-blur-md lg:col-span-2 space-y-5">
            <div className="flex items-center justify-between border-b border-border pb-2">
                <h3 className="font-semibold">Dokumentasi Foto</h3>
                <div className="flex items-center gap-2">
                    {isAuditor && (
                        <select
                            value={photoType}
                            onChange={(e) => setPhotoType(e.target.value as "Initial" | "FollowUp")}
                            className="h-9 rounded-xl border border-border bg-card px-3 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary"
                        >
                            <option value="Initial">Temuan Awal</option>
                            <option value="FollowUp">Follow-Up</option>
                        </select>
                    )}
                    
                    <Button
                        size="sm"
                        variant="outline"
                        isLoading={uploadMutation.isPending}
                        disabled={isUploadDisabled}
                        title={!isAuditor && !canUploadFollowUp ? "Klik 'Mulai Kerjakan' terlebih dahulu untuk mengunggah foto follow-up" : undefined}
                        onClick={() => fileInputRef.current?.click()}
                    >
                        <Upload className="mr-1.5 h-3.5 w-3.5" /> Upload
                    </Button>
                    
                    <input
                        ref={fileInputRef}
                        type="file"
                        accept="image/*"
                        className="hidden"
                        onChange={handleFileChange}
                        disabled={isUploadDisabled}
                    />
                </div>
            </div>

            {/* Progress Bar */}
            {uploadMutation.isPending && uploadProgress > 0 && (
                <div className="w-full bg-muted rounded-full h-2.5 overflow-hidden">
                    <div 
                        className="bg-primary h-2.5 rounded-full transition-all duration-300" 
                        style={{ width: `${uploadProgress}%` }}
                    ></div>
                    <p className="text-[10px] text-right mt-1 text-muted-foreground">{uploadProgress}%</p>
                </div>
            )}

            {/* Initial Photos */}
            <div>
                <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">
                    📌 Foto Temuan Awal ({initialPhotos.length})
                </p>
                {initialPhotos.length === 0 ? (
                    <div className="flex items-center justify-center h-24 rounded-2xl border border-dashed border-border text-muted-foreground text-sm">
                        <ImageIcon className="mr-2 h-4 w-4" /> Belum ada foto
                    </div>
                ) : (
                    <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                        {initialPhotos.map((photo) => (
                            <PhotoCard
                                key={photo.issue_photo_id}
                                photo={photo}
                                onPreview={() => setSelectedImage(photo.image_url)}
                                onDelete={() => deleteMutation.mutate(photo.issue_photo_id)}
                                isDeleting={deleteMutation.isPending}
                                isAuditor={isAuditor}
                                isInitialPhoto={true}
                            />
                        ))}
                    </div>
                )}
            </div>

            {/* Follow-Up Photos */}
            <div>
                <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">
                    Foto Follow-Up ({followUpPhotos.length})
                </p>
                {followUpPhotos.length === 0 ? (
                    <div className="flex items-center justify-center h-24 rounded-2xl border border-dashed border-border text-muted-foreground text-sm">
                        <ImageIcon className="mr-2 h-4 w-4" /> Belum ada foto follow-up
                    </div>
                ) : (
                    <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                        {followUpPhotos.map((photo) => (
                            <PhotoCard
                                key={photo.issue_photo_id}
                                photo={photo}
                                onPreview={() => setSelectedImage(photo.image_url)}
                                onDelete={() => deleteMutation.mutate(photo.issue_photo_id)}
                                isDeleting={deleteMutation.isPending}
                                isAuditor={isAuditor}
                                isInitialPhoto={false}
                                canDelete={isAuditor || canUploadFollowUp}
                            />
                        ))}
                    </div>
                )}
            </div>
        </Card>
    )
}