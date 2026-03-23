#!/bin/bash
sed -i 's/<Table>/<SmartTable data={content} \/>/g' ui/src/components/tools/rich-result-viewer.tsx
