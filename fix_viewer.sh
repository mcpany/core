sed -i 's/result: any;/result: unknown;/g' ui/src/components/tools/rich-result-viewer.tsx
sed -i 's/\/\* eslint-disable-next-line @typescript-eslint\/no-explicit-any \*\/ item: any/item: any/g' ui/src/components/tools/rich-result-viewer.tsx
sed -i 's/\/\* eslint-disable-next-line @typescript-eslint\/no-explicit-any \*\/ value: any/value: any/g' ui/src/components/tools/rich-result-viewer.tsx
sed -i 's/\/\* eslint-disable-next-line @typescript-eslint\/no-explicit-any \*\/ row: any/row: any/g' ui/src/components/tools/rich-result-viewer.tsx

sed -i 's/(item: any)/(\/* eslint-disable-next-line @typescript-eslint\/no-explicit-any *\/ item: any)/g' ui/src/components/tools/rich-result-viewer.tsx
sed -i 's/renderCell = (value: any)/renderCell = (\/* eslint-disable-next-line @typescript-eslint\/no-explicit-any *\/ value: any)/g' ui/src/components/tools/rich-result-viewer.tsx
sed -i 's/(row: any/( \/* eslint-disable-next-line @typescript-eslint\/no-explicit-any *\/ row: any/g' ui/src/components/tools/rich-result-viewer.tsx
