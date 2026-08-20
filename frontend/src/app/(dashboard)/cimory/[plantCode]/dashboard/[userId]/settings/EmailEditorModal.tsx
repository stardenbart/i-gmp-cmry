"use client";

import { useEffect, useRef, useState } from "react";
import dynamic from "next/dynamic";
import { X, Save } from "lucide-react";
import { Button } from "@/components/ui/button";

// Import dynamically to prevent SSR issues (window is not defined)
const EmailEditor = dynamic(() => import("react-email-editor").then(mod => mod.default || mod), { ssr: false });

interface EmailEditorModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (html: string, design: any) => void;
  isSaving: boolean;
  title: string;
  initialData?: string; // JSON string containing {html, design}
}

export function EmailEditorModal({
  isOpen,
  onClose,
  onSave,
  isSaving,
  title,
  initialData
}: EmailEditorModalProps) {
  const emailEditorRef = useRef<any>(null);
  const [editorReady, setEditorReady] = useState(false);

  // Load design when editor is ready or when initialData changes
  useEffect(() => {
    if (editorReady && initialData && emailEditorRef.current?.editor) {
      const trimmed = initialData.trim();
      if (trimmed.startsWith("{") || trimmed.startsWith("[")) {
        try {
          const parsed = JSON.parse(trimmed);
          if (parsed && parsed.design) {
            emailEditorRef.current.editor.loadDesign(parsed.design);
          }
        } catch {
          // Ignore invalid JSON silently (e.g. if initialData is raw HTML)
        }
      }
    }
  }, [editorReady, initialData]);

  const handleExport = () => {
    if (!emailEditorRef.current?.editor) return;
    
    emailEditorRef.current.editor.exportHtml((data: any) => {
      const { design, html } = data;
      onSave(html, design);
    });
  };

  const onReady = () => {
    setEditorReady(true);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[100] flex flex-col bg-background">
      {/* Header */}
      <div className="flex items-center justify-between px-6 py-4 border-b border-border bg-card">
        <div>
          <h2 className="text-xl font-bold">{title}</h2>
          <p className="text-sm text-muted-foreground">Editor visual untuk kustomisasi template email</p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="outline" onClick={onClose} disabled={isSaving}>
            <X className="w-4 h-4 mr-2" /> Tutup
          </Button>
          <Button onClick={handleExport} isLoading={isSaving}>
            <Save className="w-4 h-4 mr-2" /> Simpan Template
          </Button>
        </div>
      </div>

      {/* Editor Area */}
      <div className="flex-1 bg-muted relative">
        <div className="absolute inset-0">
          <EmailEditor 
            ref={emailEditorRef} 
            onReady={onReady} 
            style={{ minHeight: "100%", width: "100%" }}
            options={{
              features: {
                textEditor: {
                  spellChecker: true,
                },
              },
            }}
          />
        </div>
      </div>
    </div>
  );
}
