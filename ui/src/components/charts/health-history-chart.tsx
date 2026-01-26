"use client";

import { useMemo } from "react";
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis, ReferenceArea } from "recharts";
import { useTheme } from "next-themes";

interface HealthStatus {
    timestamp: string; // ISO string
    status: string;
    error?: string;
    latency: string; // Go duration string
}

interface HealthHistoryChartProps {
    data: HealthStatus[];
}

function parseDurationMs(d: string): number {
    if (!d) return 0;
    if (d.endsWith("ms")) return parseFloat(d);
    if (d.endsWith("µs") || d.endsWith("us")) return parseFloat(d) / 1000;
    if (d.endsWith("ns")) return parseFloat(d) / 1000000;
    if (d.endsWith("s")) return parseFloat(d) * 1000;
    if (d.endsWith("m")) return parseFloat(d) * 60000;
    return 0;
}

export function HealthHistoryChart({ data }: HealthHistoryChartProps) {
    const { theme } = useTheme();

    const chartData = useMemo(() => {
        return data.map(d => {
            const latencyMs = parseDurationMs(d.latency);
            return {
                ...d,
                displayTime: new Date(d.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
                latencyValue: latencyMs,
                isError: d.status !== "ok",
                errorMessage: d.error
            };
        });
    }, [data]);

    if (!data || data.length === 0) {
        return (
            <div className="flex items-center justify-center h-[200px] border rounded-md border-dashed text-muted-foreground bg-muted/20">
                No health history available
            </div>
        );
    }

    const maxLatency = Math.max(...chartData.map(d => d.latencyValue), 10);

    return (
        <div className="h-[200px] w-full mt-4">
            <div className="text-sm font-medium mb-2 flex items-center justify-between">
                <span>Health History (Last 100 checks)</span>
                <span className="text-xs text-muted-foreground">Latency (ms)</span>
            </div>
            <div className="h-[180px] w-full border rounded-md p-2 bg-background/50 backdrop-blur-sm">
                <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={chartData}>
                        <defs>
                            <linearGradient id="colorLatency" x1="0" y1="0" x2="0" y2="1">
                                <stop offset="5%" stopColor="#10b981" stopOpacity={0.3}/>
                                <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
                            </linearGradient>
                            <linearGradient id="colorError" x1="0" y1="0" x2="0" y2="1">
                                <stop offset="5%" stopColor="#ef4444" stopOpacity={0.3}/>
                                <stop offset="95%" stopColor="#ef4444" stopOpacity={0}/>
                            </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" opacity={0.1} vertical={false} />
                        <XAxis
                            dataKey="displayTime"
                            fontSize={10}
                            tickLine={false}
                            axisLine={false}
                            minTickGap={30}
                            opacity={0.5}
                        />
                        <YAxis
                            width={35}
                            fontSize={10}
                            tickLine={false}
                            axisLine={false}
                            tickFormatter={(val) => `${Math.round(val)}`}
                            domain={[0, 'auto']}
                        />
                        <Tooltip
                            content={({ active, payload, label }) => {
                                if (active && payload && payload.length) {
                                    const data = payload[0].payload;
                                    return (
                                        <div className="rounded-lg border bg-popover p-2 shadow-sm text-xs">
                                            <div className="font-bold mb-1">{label}</div>
                                            <div className="flex items-center gap-2">
                                                <div className={`w-2 h-2 rounded-full ${data.isError ? 'bg-red-500' : 'bg-green-500'}`} />
                                                <span>{data.status.toUpperCase()}</span>
                                            </div>
                                            <div>Latency: {data.latencyValue.toFixed(2)}ms</div>
                                            {data.errorMessage && (
                                                <div className="text-red-500 mt-1 max-w-[200px] break-words">
                                                    {data.errorMessage}
                                                </div>
                                            )}
                                        </div>
                                    );
                                }
                                return null;
                            }}
                        />
                        <Area
                            type="monotone"
                            dataKey="latencyValue"
                            stroke="#10b981"
                            strokeWidth={2}
                            fillOpacity={1}
                            fill="url(#colorLatency)"
                            isAnimationActive={false}
                        />
                        {/* Overlay errors */}
                        {chartData.map((entry, index) => (
                            entry.isError ? (
                                <ReferenceArea
                                    key={index}
                                    x1={entry.displayTime}
                                    x2={entry.displayTime}
                                    strokeOpacity={0.3}
                                    fill="red"
                                    fillOpacity={0.1}
                                />
                            ) : null
                        ))}
                    </AreaChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
}
