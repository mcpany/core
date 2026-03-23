const fs = require('fs');

let content = fs.readFileSync('ui/src/components/stats/health-history-chart.tsx', 'utf-8');

const regex = /const fetchData = async \(\) => \{[\s\S]*?fetchData\(\);/m;

const newFetchData = `
        const fetchData = async () => {
            try {
                // Fetch real health data from the dashboard/health endpoint
                const healthData = await apiClient.getDashboardHealth();

                const points: HealthPoint[] = [];
                const historyMap = healthData.history || {};

                // Collect all points to find min and max time
                let minTime = Infinity;
                let maxTime = -Infinity;
                const allPoints: { timestamp: number, status: string }[] = [];

                Object.values(historyMap).forEach((serviceHistory) => {
                    if (Array.isArray(serviceHistory)) {
                        serviceHistory.forEach((point) => {
                            if (point && point.timestamp) {
                                minTime = Math.min(minTime, point.timestamp);
                                maxTime = Math.max(maxTime, point.timestamp);
                                allPoints.push(point);
                            }
                        });
                    }
                });

                if (allPoints.length === 0) {
                    setData([]);
                    return;
                }

                // Bucket the timestamps into chunks (e.g. 60 buckets)
                const buckets = 60;
                let timeSpan = maxTime - minTime;

                // If the timespan is too small (e.g. 0), use a default 1-minute bucket length
                if (timeSpan <= 0) {
                    timeSpan = 60000;
                    maxTime = minTime + timeSpan;
                }
                const bucketSize = timeSpan / buckets;

                const bucketData: Record<number, { up: number, total: number }> = {};
                for (let i = 0; i < buckets; i++) {
                    bucketData[i] = { up: 0, total: 0 };
                }

                Object.values(historyMap).forEach((serviceHistory) => {
                    if (Array.isArray(serviceHistory)) {
                        serviceHistory.forEach((point) => {
                            if (point && point.timestamp) {
                                let bIndex = Math.floor((point.timestamp - minTime) / bucketSize);
                                bIndex = Math.min(Math.max(bIndex, 0), buckets - 1); // clamp to avoid out-of-bounds
                                bucketData[bIndex].total += 1;
                                // Assume healthy if status is "healthy", "up", "ok", etc.
                                if (["healthy", "up", "ok", "serving"].includes(point.status.toLowerCase())) {
                                    bucketData[bIndex].up += 1;
                                }
                            }
                        });
                    }
                });

                // Convert buckets back to array
                for (let i = 0; i < buckets; i++) {
                    const bucket = bucketData[i];
                    // Create a time label for the bucket
                    const time = new Date(minTime + i * bucketSize).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

                    if (bucket && bucket.total > 0) {
                        const uptimeRatio = bucket.up / bucket.total;
                        let status: HealthPoint["status"] = "ok";
                        if (uptimeRatio < 0.5) status = "error";
                        else if (uptimeRatio < 1.0) status = "degraded";

                        points.push({
                            time,
                            status,
                            uptime: Math.round(uptimeRatio * 100)
                        });
                    } else {
                        // If no data in bucket, fill with previous state or 100% ok
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

        fetchData();`;

content = content.replace(regex, newFetchData);

fs.writeFileSync('ui/src/components/stats/health-history-chart.tsx', content);
