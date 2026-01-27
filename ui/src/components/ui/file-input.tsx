/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import * as React from "react"
import { Upload, X, FileText } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import {
  Dialog,
  DialogContent,
  DialogTrigger,
} from "@/components/ui/dialog"

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
       if (!previewUrl && accept?.includes("image")) {
           setPreviewUrl(`data:image/png;base64,${value}`)
       }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value, accept])

  const [isDragging, setIsDragging] = React.useState(false)

  const processFile = (file: File) => {
    setError(null)

    // Check size (optional, e.g. 5MB limit to prevent browser crash)
    if (file.size > 5 * 1024 * 1024) {
      setError("File is too large (max 5MB)")
      return
    }

    setFileName(file.name)

    const reader = new FileReader()
    reader.onload = (event) => {
      const result = event.target?.result as string
      // Check if it's an image for preview
      if (file.type.startsWith("image/")) {
          setPreviewUrl(result)
      } else {
          setPreviewUrl(null)
      }
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

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    e.stopPropagation()
    if (!disabled) {
      setIsDragging(true)
    }
  }

  const handleDragLeave = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)
  }

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
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
    <div
      className={cn(
        "flex flex-col gap-2 rounded-md border border-transparent transition-all",
        isDragging && "border-dashed border-primary bg-primary/5 p-4",
        className
      )}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <input
        type="file"
        id={id}
        ref={inputRef}
        onChange={handleFileChange}
        accept={accept}
        className="hidden"
        disabled={disabled}
      />
      <div className="flex items-center gap-2">
        {previewUrl && (
            <Dialog>
                <DialogTrigger asChild>
                    <div className="relative group cursor-zoom-in shrink-0">
                        {/* eslint-disable-next-line @next/next/no-img-element */}
                        <img
                            src={previewUrl}
                            alt="Preview"
                            className="h-9 w-9 rounded object-cover border border-border bg-muted/50 shadow-sm"
                        />
                        <div className="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors rounded" />
                    </div>
                </DialogTrigger>
                <DialogContent className="max-w-[90vw] max-h-[90vh] p-0 overflow-hidden bg-transparent border-0 shadow-none flex justify-center items-center">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img
                        src={previewUrl}
                        alt="Preview"
                        className="max-h-[85vh] max-w-full rounded-lg shadow-2xl"
                    />
                </DialogContent>
            </Dialog>
        )}

        <Button
          type="button"
          variant="outline"
          onClick={() => inputRef.current?.click()}
          disabled={disabled}
          className="w-full sm:w-auto"
        >
          <Upload className="mr-2 h-4 w-4" />
          {fileName ? "Change File" : "Select File"}
        </Button>
        {fileName && (
           <div className="flex items-center gap-2 bg-muted px-3 py-2 rounded-md text-sm flex-1 overflow-hidden">
             {!previewUrl && <FileText className="h-4 w-4 shrink-0" />}
             <span className="truncate">{fileName}</span>
             <Button
               type="button"
               variant="ghost"
               size="icon"
               className="h-6 w-6 ml-auto shrink-0"
               onClick={clearFile}
               disabled={disabled}
             >
               <X className="h-4 w-4" />
             </Button>
           </div>
        )}
         {isDragging && !fileName && (
            <div className="flex-1 flex items-center justify-center text-sm text-primary font-medium animate-pulse">
                Drop file here
            </div>
        )}
      </div>
      {error && <p className="text-xs text-destructive">{error}</p>}
      {accept && <p className="text-[10px] text-muted-foreground">Accepted formats: {accept}</p>}
    </div>
  )
}
