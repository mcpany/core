/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import * as React from "react"
import { Upload, X, FileText, Image as ImageIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

interface FileInputProps {
  value?: string
  onChange: (value: string | undefined) => void
  accept?: string
  className?: string
  disabled?: boolean
  id?: string
}

/**
 * FileInput is a form component for selecting a file and converting it to a Base64 string.
 * It displays the selected filename and allows clearing the selection.
 * @param props - Component props.
 * @returns The FileInput component.
 */
export function FileInput({ value, onChange, accept, className, disabled, id }: FileInputProps) {
  const inputRef = React.useRef<HTMLInputElement>(null)
  const [fileName, setFileName] = React.useState<string | null>(null)
  const [error, setError] = React.useState<string | null>(null)
  const [isDragging, setIsDragging] = React.useState(false)
  const [previewUrl, setPreviewUrl] = React.useState<string | null>(null)

  // Sync internal state with external value
  React.useEffect(() => {
    if (!value) {
      setFileName(null)
      setPreviewUrl(null)
      if (inputRef.current) {
        inputRef.current.value = ""
      }
    } else if (!fileName) {
       // Value exists but no filename (e.g. form preset loaded)
       setFileName("File loaded")
    }
  }, [value])

  // Cleanup preview URL on unmount or change
  React.useEffect(() => {
    return () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl)
      }
    }
  }, [previewUrl])

  const processFile = (file: File) => {
    setError(null)

    // Check size (optional, e.g. 5MB limit to prevent browser crash)
    if (file.size > 5 * 1024 * 1024) {
      setError("File is too large (max 5MB)")
      return
    }

    setFileName(file.name)

    // Create preview if image
    if (file.type.startsWith("image/")) {
        const url = URL.createObjectURL(file)
        setPreviewUrl(url)
    } else {
        setPreviewUrl(null)
    }

    const reader = new FileReader()
    reader.onload = (event) => {
      const result = event.target?.result as string
      // result is "data:image/png;base64,....."
      // We want only the base64 part for contentEncoding="base64"
      const base64 = result.split(",")[1]
      onChange(base64)
    }
    reader.onerror = () => {
      setError("Failed to read file")
    }
    reader.readAsDataURL(file)
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
        processFile(file)
    }
  }

  const handleDragOver = (e: React.DragEvent) => {
      e.preventDefault()
      e.stopPropagation()
      if (!disabled) {
          setIsDragging(true)
      }
  }

  const handleDragLeave = (e: React.DragEvent) => {
      e.preventDefault()
      e.stopPropagation()
      setIsDragging(false)
  }

  const handleDrop = (e: React.DragEvent) => {
      e.preventDefault()
      e.stopPropagation()
      setIsDragging(false)

      if (disabled) return

      const file = e.dataTransfer.files?.[0]
      if (file) {
          processFile(file)
      }
  }

  const clearFile = () => {
    setFileName(null)
    setPreviewUrl(null)
    onChange(undefined)
    if (inputRef.current) {
      inputRef.current.value = ""
    }
  }

  return (
    <div className={cn("flex flex-col gap-2", className)}>
      <input
        type="file"
        id={id}
        ref={inputRef}
        onChange={handleFileChange}
        accept={accept}
        className="hidden"
        disabled={disabled}
      />

      <div
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        className={cn(
            "flex flex-col items-center justify-center p-6 border-2 border-dashed rounded-lg transition-all duration-200 gap-4",
            isDragging ? "border-primary bg-primary/5" : "border-muted-foreground/25 hover:border-primary/50 hover:bg-muted/50",
            disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer",
            previewUrl ? "border-solid border-primary/20 bg-muted/20" : ""
        )}
        onClick={() => !disabled && inputRef.current?.click()}
        role="button"
        tabIndex={disabled ? -1 : 0}
        onKeyDown={(e) => {
            if (!disabled && (e.key === "Enter" || e.key === " ")) {
                e.preventDefault();
                inputRef.current?.click();
            }
        }}
        aria-label={fileName ? `Change file, currently selected: ${fileName}` : "Upload file"}
      >
        {previewUrl ? (
            <div className="relative group w-full flex justify-center">
                <img src={previewUrl} alt="Preview" className="max-h-[200px] rounded shadow-sm object-contain" />
                <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center rounded">
                    <p className="text-white font-medium text-sm">Click to change</p>
                </div>
            </div>
        ) : (
             <div className="flex flex-col items-center gap-2 text-center">
                <div className={cn("p-3 rounded-full bg-muted transition-colors", isDragging ? "bg-primary/20 text-primary" : "text-muted-foreground")}>
                    <Upload className="h-6 w-6" />
                </div>
                <div className="space-y-1">
                    <p className="text-sm font-medium text-foreground">
                        {fileName ? fileName : "Click or drag file to upload"}
                    </p>
                    <p className="text-xs text-muted-foreground">
                        {accept ? `Accepted formats: ${accept}` : "All files supported"} (max 5MB)
                    </p>
                </div>
            </div>
        )}
      </div>

      {fileName && (
           <div className="flex items-center gap-2 bg-muted px-3 py-2 rounded-md text-sm overflow-hidden border border-border/50">
             {previewUrl ? <ImageIcon className="h-4 w-4 shrink-0 text-primary" /> : <FileText className="h-4 w-4 shrink-0" />}
             <span className="truncate flex-1 font-mono text-xs">{fileName}</span>
             <Button
               type="button"
               variant="ghost"
               size="icon"
               className="h-6 w-6 shrink-0 hover:text-destructive"
               onClick={(e) => {
                   e.stopPropagation()
                   clearFile()
               }}
               disabled={disabled}
               title="Clear file"
             >
               <X className="h-4 w-4" />
             </Button>
           </div>
        )}

      {error && <p className="text-xs text-destructive flex items-center gap-1"><X className="h-3 w-3" /> {error}</p>}
    </div>
  )
}
