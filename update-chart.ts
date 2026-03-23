import fs from 'fs';

let content = fs.readFileSync('ui/src/components/stats/health-history-chart.tsx', 'utf-8');

const newFetchData = `
        const fetchData = async () => {
            try {
                // Fetch real health data from the dashboard/health endpoint
                const healthData = await apiClient.getDashboardHealth();

                // Get history across all services
                const points: HealthPoint[] = [];
                const historyMap = healthData.history || {};

                // Find the earliest and latest timestamps across all services
                let minTime = Infinity;
                let maxTime = -Infinity;
                const allPoints: { timestamp: number, status: string }[] = [];

                Object.values(historyMap).forEach((serviceHistory) => {
                    serviceHistory.forEach((point) => {
                        minTime = Math.min(minTime, point.timestamp);
                        maxTime = Math.max(maxTime, point.timestamp);
                        allPoints.push(point);
                    });
                });

                if (allPoints.length === 0) {
                    setData([]);
                    return;
                }

                // Bucket the timestamps into chunks (e.g. 60 buckets)
                const buckets = 60;
                const timeSpan = maxTime - minTime;
                // If the timespan is too small, use a default 1-minute bucket length
                const bucketSize = timeSpan > 0 ? timeSpan / buckets : 60000;

                const bucketData: Record<number, { up: number, total: number }> = {};

                Object.values(historyMap).forEach((serviceHistory) => {
                    serviceHistory.forEach((point) => {
                        const bIndex = Math.floor((point.timestamp - minTime) / bucketSize);
                        const bucket = Math.min(bIndex, buckets - 1); // clamp to avoid out-of-bounds
                        if (!bucketData[bucket]) {
                            bucketData[bucket] = { up: 0, total: 0 };
                        }
                        bucketData[bucket].total += 1;
                        if (point.status === "healthy" || point.status === "up" || point.status === "ok") {
                            bucketData[bucket].up += 1;
                        }
                    });
                });

                // Convert buckets back to array
                for (let i = 0; i < buckets; i++) {
                    const bucket = bucketData[i];
                    const time = new Date(minTime + i * bucketSize).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

                    if (bucket) {
                        const uptimeRatio = bucket.total > 0 ? bucket.up / bucket.total : 0;
                        let status: HealthPoint["status"] = "ok";
                        if (uptimeRatio < 0.5) status = "error";
                        else if (uptimeRatio < 1.0) status = "degraded";

                        points.push({
                            time,
                            status,
                            uptime: Math.round(uptimeRatio * 100)
                        });
                    } else {
                        // Fill empty gaps with previous state or 100% ok
                        points.push({
                            time,
                            status: "ok",
                            uptime: 100
                        });
                    }
                }

                setData(points);
            } catch (error) {
                console.error("Failed to fetch health history", error);
            } finally {
                setLoading(false);
            }
        };
`;

content = content.replace(/const fetchData = async \(\) => \{[\s\S]*?fetchData\(\);/m, newFetchData + '\n        fetchData();');
fs.writeFileSync('ui/src/components/stats/health-history-chart.tsx', content);
