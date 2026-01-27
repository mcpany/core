/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import * as React from "react"
import { Upload, X, FileText } from "lucide-react"
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

  // Cleanup object URL on unmount
  React.useEffect(() => {
    return () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl)
      }
    }
  }, [previewUrl])

  // Sync internal state with external value
  React.useEffect(() => {
    if (!value) {
      setFileName(null)
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl)
        setPreviewUrl(null)
      }
      if (inputRef.current) {
        inputRef.current.value = ""
      }
    } else if (!fileName) {
       // Value exists but no filename (e.g. form preset loaded)
       setFileName("File loaded")
    }
  }, [value])

  const [isDragging, setIsDragging] = React.useState(false)

  const processFile = (file: File) => {
    setError(null)

    // Check size (optional, e.g. 5MB limit to prevent browser crash)
    if (file.size > 5 * 1024 * 1024) {
      setError("File is too large (max 5MB)")
      return
    }

    setFileName(file.name)

    // Generate preview if it's an image
    if (file.type.startsWith("image/")) {
      const url = URL.createObjectURL(file)
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl)
      }
      setPreviewUrl(url)
    } else {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl)
        setPreviewUrl(null)
      }
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
    if (previewUrl) {
      URL.revokeObjectURL(previewUrl)
      setPreviewUrl(null)
    }
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
             <FileText className="h-4 w-4 shrink-0" />
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
      {previewUrl && (
        <div className="relative mt-2 rounded-md overflow-hidden border bg-muted/50 max-w-[200px]">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={previewUrl} alt="Preview" className="w-full h-auto object-contain max-h-48" />
        </div>
      )}
      {accept && <p className="text-[10px] text-muted-foreground">Accepted formats: {accept}</p>}
    </div>
  )
}
