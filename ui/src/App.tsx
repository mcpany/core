/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { lazy, Suspense } from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import { Layout } from "./components/layout";
import { Loader2 } from "lucide-react";

// Lazy-load every page so the initial bundle stays small.
const DashboardPage = lazy(() => import("./app/page"));
const AlertsPage = lazy(() => import("./app/alerts/page"));
const AuditPage = lazy(() => import("./app/audit/page"));
const AuthCallbackPage = lazy(() => import("./app/auth/callback/page"));
const ConfigValidatorPage = lazy(() => import("./app/config-validator/page"));
const ContextPage = lazy(() => import("./app/context/page"));
const CredentialsPage = lazy(() => import("./app/credentials/page"));
const DiagnosticsPage = lazy(() => import("./app/diagnostics/page"));
const InspectorPage = lazy(() => import("./app/inspector/page"));
const LoginPage = lazy(() => import("./app/login/page"));
const LogsPage = lazy(() => import("./app/logs/page"));
const MarketplacePage = lazy(() => import("./app/marketplace/page"));
const ExternalMarketplacePage = lazy(() => import("./app/marketplace/external/[id]/page"));
const MiddlewarePage = lazy(() => import("./app/middleware/page"));
const NetworkPage = lazy(() => import("./app/network/page"));
const OAuthCallbackPage = lazy(() => import("./app/oauth/callback/page"));
const PlaygroundPage = lazy(() => import("./app/playground/page"));
const PlaygroundSchemaPage = lazy(() => import("./app/playground/schema/page"));
const ProfilesPage = lazy(() => import("./app/profiles/page"));
const PromptsPage = lazy(() => import("./app/prompts/page"));
const ResourcesPage = lazy(() => import("./app/resources/page"));
const SecretsPage = lazy(() => import("./app/secrets/page"));
const ServiceDetailPage = lazy(() => import("./app/service/[id]/page"));
const ServicePromptPage = lazy(() => import("./app/service/[id]/prompt/[name]/page"));
const ServiceResourcePage = lazy(() => import("./app/service/[id]/resource/[name]/page"));
const ServiceToolPage = lazy(() => import("./app/service/[id]/tool/[name]/page"));
const SettingsMiddlewarePage = lazy(() => import("./app/settings/middleware/page"));
const SettingsPage = lazy(() => import("./app/settings/page"));
const HitlPage = lazy(() => import("./app/hitl/page"));
const BlackboardPage = lazy(() => import("./app/blackboard/page"));
const SettingsWebhooksPage = lazy(() => import("./app/settings/webhooks/page"));
const SkillEditPage = lazy(() => import("./app/skills/[name]/edit/page"));
const SkillNamePage = lazy(() => import("./app/skills/[name]/page"));
const SkillCreatePage = lazy(() => import("./app/skills/create/page"));
const SkillsPage = lazy(() => import("./app/skills/page"));
const StackDetailPage = lazy(() => import("./app/stacks/[stackId]/page"));
const StacksPage = lazy(() => import("./app/stacks/page"));
const StatsPage = lazy(() => import("./app/stats/page"));
const ToolsPage = lazy(() => import("./app/tools/page"));
const TracesPage = lazy(() => import("./app/traces/page"));
const UpstreamServiceDetailPage = lazy(() => import("./app/upstream-services/[serviceId]/page"));
const UpstreamServicesPage = lazy(() => import("./app/upstream-services/page"));
const UsersPage = lazy(() => import("./app/users/page"));
const VisualizerPage = lazy(() => import("./app/visualizer/page"));
const WebhooksPage = lazy(() => import("./app/webhooks/page"));
const UniversalAgentBusPage = lazy(() => import("./app/universal-agent-bus/page"));

const PageFallback = () => (
  <div className="flex items-center justify-center h-full min-h-[200px]">
    <Loader2 className="animate-spin h-8 w-8 text-muted-foreground" />
  </div>
);

/**
 * Intent: Document App
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * Root application component with React Router routes.
 * Public routes (login, auth) get a top-level Suspense fallback.
 * Protected routes (wrapped in Layout) have Suspense inside the Layout so the
 * sidebar/header stay visible while the page content lazy-loads.
 */
export default function App() {
  return (
    <Suspense fallback={<PageFallback />}>
      <Routes>
        {/* Public routes (no sidebar) */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/auth/callback" element={<AuthCallbackPage />} />
        <Route path="/oauth/callback" element={<OAuthCallbackPage />} />

        {/* Main application routes (wrapped in the sidebar layout) */}
        <Route element={<Layout />}>
          <Route index element={<DashboardPage />} />
          <Route path="/alerts" element={<AlertsPage />} />
          <Route path="/audit" element={<AuditPage />} />
          <Route path="/config-validator" element={<ConfigValidatorPage />} />
          <Route path="/context" element={<ContextPage />} />
          <Route path="/hitl" element={<HitlPage />} />
          <Route path="/blackboard" element={<BlackboardPage />} />
          <Route path="/credentials" element={<CredentialsPage />} />
          <Route path="/diagnostics" element={<DiagnosticsPage />} />
          <Route path="/inspector" element={<InspectorPage />} />
          <Route path="/logs" element={<LogsPage />} />
          <Route path="/marketplace" element={<MarketplacePage />} />
          <Route path="/marketplace/external/:id" element={<ExternalMarketplacePage />} />
          <Route path="/middleware" element={<MiddlewarePage />} />
          <Route path="/network" element={<NetworkPage />} />
          <Route path="/playground" element={<PlaygroundPage />} />
          <Route path="/playground/schema" element={<PlaygroundSchemaPage />} />
          <Route path="/profiles" element={<ProfilesPage />} />
          <Route path="/prompts" element={<PromptsPage />} />
          <Route path="/resources" element={<ResourcesPage />} />
          <Route path="/secrets" element={<SecretsPage />} />
          <Route path="/service/:id" element={<ServiceDetailPage />} />
          <Route path="/service/:id/tool/:name" element={<ServiceToolPage />} />
          <Route path="/service/:id/resource/:name" element={<ServiceResourcePage />} />
          <Route path="/service/:id/prompt/:name" element={<ServicePromptPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/settings/middleware" element={<SettingsMiddlewarePage />} />
          <Route path="/settings/webhooks" element={<SettingsWebhooksPage />} />
          <Route path="/skills" element={<SkillsPage />} />
          <Route path="/skills/create" element={<SkillCreatePage />} />
          <Route path="/skills/:name" element={<SkillNamePage />} />
          <Route path="/skills/:name/edit" element={<SkillEditPage />} />
          <Route path="/stacks" element={<StacksPage />} />
          <Route path="/stacks/:stackId" element={<StackDetailPage />} />
          <Route path="/stats" element={<StatsPage />} />
          <Route path="/tools" element={<ToolsPage />} />
          <Route path="/traces" element={<TracesPage />} />
          <Route path="/upstream-services" element={<UpstreamServicesPage />} />
          <Route path="/upstream-services/:serviceId" element={<UpstreamServiceDetailPage />} />
          <Route path="/users" element={<UsersPage />} />
          <Route path="/visualizer" element={<VisualizerPage />} />
          <Route path="/webhooks" element={<WebhooksPage />} />
          <Route path="/universal-agent-bus" element={<UniversalAgentBusPage />} />

          {/* Legacy redirect */}
          <Route path="/topology" element={<Navigate to="/network" replace />} />
        </Route>
      </Routes>
    </Suspense>
  );
}
