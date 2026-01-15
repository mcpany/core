/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback } from 'react';
import { apiClient, DoctorReport } from '@/lib/client';

export function useSystemHealth(pollingInterval = 30000) {
  const [report, setReport] = useState<DoctorReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchHealth = useCallback(async () => {
    try {
      setLoading(true);
      const data = await apiClient.getDoctorReport();
      setReport(data);
      setError(null);
    } catch (err) {
      setError(err as Error);
      setReport(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchHealth();
    const interval = setInterval(fetchHealth, pollingInterval);
    return () => clearInterval(interval);
  }, [fetchHealth, pollingInterval]);

  return { report, loading, error, refresh: fetchHealth };
}
