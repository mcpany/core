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
       // Check if we can show a preview from base64
       if (value.length > 0) {
           // We try to guess if it's an image.
           // If accept starts with image/, we assume it is.
           if (accept && accept.startsWith("image/")) {
               // We need a MIME type. If accept is specific (image/png), use it.
               // If generic (image/*), default to png or jpeg?
               const mime = accept === "image/*" ? "image/png" : accept;
               setPreviewUrl(`data:${mime};base64,${value}`)
           }
       }
    }
  }, [value, accept]) // Added accept to deps

  // Cleanup object URL
  React.useEffect(() => {
    return () => {
      if (previewUrl && previewUrl.startsWith("blob:")) {
        URL.revokeObjectURL(previewUrl)
      }
    }
  }, [previewUrl])

  const [isDragging, setIsDragging] = React.useState(false)

  const processFile = (file: File) => {
    setError(null)

    // Check size (optional, e.g. 5MB limit to prevent browser crash)
    if (file.size > 5 * 1024 * 1024) {
      setError("File is too large (max 5MB)")
      return
    }

    setFileName(file.name)

    // Generate preview if image
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
      <div className="flex items-start gap-2">
        <Button
          type="button"
          variant="outline"
          onClick={() => inputRef.current?.click()}
          disabled={disabled}
          className="w-full sm:w-auto shrink-0"
        >
          <Upload className="mr-2 h-4 w-4" />
          {fileName ? "Change File" : "Select File"}
        </Button>

        {fileName && (
           <div className="flex flex-col gap-2 flex-1 min-w-0 bg-muted px-3 py-2 rounded-md text-sm">
             <div className="flex items-center gap-2 overflow-hidden">
                 {previewUrl ? (
                    <ImageIcon className="h-4 w-4 shrink-0 text-primary" />
                 ) : (
                    <FileText className="h-4 w-4 shrink-0" />
                 )}
                 <span className="truncate flex-1">{fileName}</span>
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
             {previewUrl && (
                 <div className="relative w-full h-32 bg-background/50 rounded overflow-hidden border">
                     {/* eslint-disable-next-line @next/next/no-img-element */}
                     <img
                        src={previewUrl}
                        alt="Preview"
                        className="w-full h-full object-contain"
                     />
                 </div>
             )}
           </div>
        )}

         {isDragging && !fileName && (
            <div className="flex-1 flex items-center justify-center text-sm text-primary font-medium animate-pulse border border-dashed rounded-md p-2">
                Drop file here
            </div>
        )}
      </div>
      {error && <p className="text-xs text-destructive">{error}</p>}
      {accept && <p className="text-[10px] text-muted-foreground">Accepted formats: {accept}</p>}
    </div>
  )
}
