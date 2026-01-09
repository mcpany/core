export type Severity = "critical" | "warning" | "info";
export type AlertStatus = "active" | "acknowledged" | "resolved";

export interface Alert {
  id: string;
  title: string;
  message: string;
  severity: Severity;
  status: AlertStatus;
  service: string;
  timestamp: string; // ISO string
  source: string;
}

export interface AlertRule {
  id: string;
  name: string;
  condition: string;
  severity: Severity;
  service: string;
  enabled: boolean;
}
