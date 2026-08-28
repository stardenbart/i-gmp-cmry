"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import dynamic from "next/dynamic";
import { X, Save } from "lucide-react";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import type { EditorRef } from "react-email-editor";
import type { JSONTemplate } from "@unlayer/types";

// Import dynamically to prevent SSR issues (window is not defined)
const EmailEditor = dynamic(() => import("react-email-editor").then(mod => mod.default || mod), { ssr: false });

interface EmailEditorModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (html: string, design: unknown) => void;
  isSaving: boolean;
  title: string;
  initialData?: string; // JSON string containing {html, design}
  variables?: readonly string[];
  requiredVariables?: readonly string[];
}

// Wraps raw HTML (from a legacy/seeded plain-HTML template) in a minimal
// Unlayer design so it opens in the visual editor instead of a code box.
// The HTML lands in a single "html" content block, which Unlayer renders
// and lets the user edit/rearrange like any other block.
function designFromLegacyHTML(html: string): JSONTemplate<"email"> {
  return {
    counters: {},
    body: {
      rows: [
        {
          cells: [1],
          columns: [
            {
              contents: [
                {
                  type: "html",
                  values: {
                    html,
                    synced: { id: "legacy-html", dirty: false },
                  },
                },
              ],
              values: {},
            },
          ],
          values: {},
        },
      ],
      values: {},
    },
    schemaVersion: 12,
  } as unknown as JSONTemplate<"email">;
}

export function EmailEditorModal({
  isOpen,
  onClose,
  onSave,
  isSaving,
  title,
  initialData,
  variables = [],
  requiredVariables = []
}: EmailEditorModalProps) {
  const emailEditorRef = useRef<EditorRef>(null);
  const [editorReady, setEditorReady] = useState(false);
  const designToLoad = useMemo(() => {
    const trimmed = initialData?.trim() || "";
    if (trimmed.startsWith("{") || trimmed.startsWith("[")) {
      try {
        const parsed = JSON.parse(trimmed) as {
          html?: unknown;
          design?: JSONTemplate<"email">;
        };
        if (parsed?.design) return parsed.design;
        if (typeof parsed?.html === "string") return designFromLegacyHTML(parsed.html);
      } catch {
        // Treat malformed JSON as legacy template content.
      }
    }
    // Not JSON at all: either a raw legacy HTML string, or empty (new template).
    return trimmed ? designFromLegacyHTML(initialData || "") : null;
  }, [initialData]);

  // Load design when editor is ready or when initialData changes
  useEffect(() => {
    if (editorReady && designToLoad && emailEditorRef.current?.editor) {
      emailEditorRef.current.editor.loadDesign(designToLoad);
    }
  }, [editorReady, designToLoad]);

  const handleExport = () => {
    if (!emailEditorRef.current?.editor) return;

    emailEditorRef.current.editor.exportHtml((data) => {
      const { design, html } = data;
      const missingVariables = requiredVariables.filter((variable) => !html.includes(`{{.${variable}}}`));
      if (missingVariables.length > 0) {
        toast.error(`Template wajib memuat: ${missingVariables.map((variable) => `{{.${variable}}}`).join(", ")}`);
        return;
      }
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
          <p className="text-sm text-muted-foreground">
            Editor visual untuk kustomisasi template email
          </p>
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
        {variables.length > 0 && (
          <div className="absolute left-4 right-4 top-3 z-10 rounded-lg border border-border bg-background/95 px-3 py-2 text-xs shadow-sm">
            Variabel tersedia: {variables.map((variable) => `{{.${variable}}}`).join(", ")}
            {requiredVariables.length > 0 && (
              <strong className="ml-2 text-destructive">
                Wajib: {requiredVariables.map((variable) => `{{.${variable}}}`).join(", ")}
              </strong>
            )}
          </div>
        )}
        <div className={`absolute inset-x-0 bottom-0 ${variables.length > 0 ? "top-14" : "top-0"}`}>
          <EmailEditor
            ref={emailEditorRef}
            onReady={onReady}
            style={{ minHeight: "100%", width: "100%" }}
            options={{
              mergeTags: Object.fromEntries(variables.map((variable) => [variable, { name: variable, value: `{{.${variable}}}` }])),
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
