sed -i 's/let wsSend: ((data: string) => void) | null = null;/let wsSend: any = null;/g' ui/tests/inspector.spec.ts
