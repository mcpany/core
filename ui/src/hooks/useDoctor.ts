/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from 'react';
import { DoctorReport } from '@/types/doctor';

export function useDoctor(pollingInterval = 10000) {
  const [report, setReport] = useState<DoctorReport | null>(null);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchDoctor = async () => {
      try {
        const res = await fetch('/api/v1/doctor');
        if (!res.ok) {
          throw new Error(`Failed to fetch doctor report: ${res.statusText}`);
        }
        const data: DoctorReport = await res.json();
        setReport(data);
        setError(null);
      } catch (err) {
        setError(err as Error);
      }
    };

    fetchDoctor();
    const interval = setInterval(fetchDoctor, pollingInterval);

    return () => clearInterval(interval);
  }, [pollingInterval]);

  return { report, error };
}
