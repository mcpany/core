const fs = require('fs');
let file = fs.readFileSync('ui/src/components/stats/health-history-chart.tsx', 'utf-8');
file = file.replace(/import \{ useState, useEffect, useMemo \} from "react";/, 'import { useState, useEffect } from "react";');
file = file.replace(/const \[loading, setLoading\] = useState\(true\);/, '');
file = file.replace(/setLoading\(false\);/g, '');
fs.writeFileSync('ui/src/components/stats/health-history-chart.tsx', file);
