/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { lazy, Suspense } from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import { Layout } from "./components/layout";
import { Loader2 } from "lucide-react";

// Lazy-load every page so the initial bundle stays small.
/**
 * Summary: DashboardPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const DashboardPage = lazy(() => import("./app/page"));
/**
 * Summary: AlertsPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const AlertsPage = lazy(() => import("./app/alerts/page"));
/**
 * Summary: AuditPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const AuditPage = lazy(() => import("./app/audit/page"));
/**
 * Summary: AuthCallbackPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const AuthCallbackPage = lazy(() => import("./app/auth/callback/page"));
/**
 * Summary: ConfigValidatorPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const ConfigValidatorPage = lazy(() => import("./app/config-validator/page"));
/**
 * Summary: ContextPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const ContextPage = lazy(() => import("./app/context/page"));
/**
 * Summary: CredentialsPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const CredentialsPage = lazy(() => import("./app/credentials/page"));
/**
 * Summary: DiagnosticsPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const DiagnosticsPage = lazy(() => import("./app/diagnostics/page"));
/**
 * Summary: InspectorPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const InspectorPage = lazy(() => import("./app/inspector/page"));
/**
 * Summary: LoginPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const LoginPage = lazy(() => import("./app/login/page"));
/**
 * Summary: LogsPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const LogsPage = lazy(() => import("./app/logs/page"));
/**
 * Summary: MarketplacePage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const MarketplacePage = lazy(() => import("./app/marketplace/page"));
/**
 * Summary: ExternalMarketplacePage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const ExternalMarketplacePage = lazy(() => import("./app/marketplace/external/[id]/page"));
/**
 * Summary: MiddlewarePage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const MiddlewarePage = lazy(() => import("./app/middleware/page"));
/**
 * Summary: NetworkPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const NetworkPage = lazy(() => import("./app/network/page"));
/**
 * Summary: OAuthCallbackPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const OAuthCallbackPage = lazy(() => import("./app/oauth/callback/page"));
/**
 * Summary: PlaygroundPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const PlaygroundPage = lazy(() => import("./app/playground/page"));
/**
 * Summary: PlaygroundSchemaPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const PlaygroundSchemaPage = lazy(() => import("./app/playground/schema/page"));
/**
 * Summary: ProfilesPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const ProfilesPage = lazy(() => import("./app/profiles/page"));
/**
 * Summary: PromptsPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const PromptsPage = lazy(() => import("./app/prompts/page"));
/**
 * Summary: ResourcesPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const ResourcesPage = lazy(() => import("./app/resources/page"));
/**
 * Summary: SecretsPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const SecretsPage = lazy(() => import("./app/secrets/page"));
/**
 * Summary: ServiceDetailPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const ServiceDetailPage = lazy(() => import("./app/service/[id]/page"));
/**
 * Summary: ServicePromptPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const ServicePromptPage = lazy(() => import("./app/service/[id]/prompt/[name]/page"));
/**
 * Summary: ServiceResourcePage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const ServiceResourcePage = lazy(() => import("./app/service/[id]/resource/[name]/page"));
/**
 * Summary: ServiceToolPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const ServiceToolPage = lazy(() => import("./app/service/[id]/tool/[name]/page"));
/**
 * Summary: SettingsMiddlewarePage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const SettingsMiddlewarePage = lazy(() => import("./app/settings/middleware/page"));
/**
 * Summary: SettingsPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const SettingsPage = lazy(() => import("./app/settings/page"));
/**
 * Summary: SettingsWebhooksPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const SettingsWebhooksPage = lazy(() => import("./app/settings/webhooks/page"));
/**
 * Summary: SkillEditPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const SkillEditPage = lazy(() => import("./app/skills/[name]/edit/page"));
/**
 * Summary: SkillNamePage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const SkillNamePage = lazy(() => import("./app/skills/[name]/page"));
/**
 * Summary: SkillCreatePage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const SkillCreatePage = lazy(() => import("./app/skills/create/page"));
/**
 * Summary: SkillsPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const SkillsPage = lazy(() => import("./app/skills/page"));
/**
 * Summary: StackDetailPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const StackDetailPage = lazy(() => import("./app/stacks/[stackId]/page"));
/**
 * Summary: StacksPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const StacksPage = lazy(() => import("./app/stacks/page"));
/**
 * Summary: StatsPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const StatsPage = lazy(() => import("./app/stats/page"));
/**
 * Summary: ToolsPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const ToolsPage = lazy(() => import("./app/tools/page"));
/**
 * Summary: TracesPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const TracesPage = lazy(() => import("./app/traces/page"));
/**
 * Summary: UpstreamServiceDetailPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const UpstreamServiceDetailPage = lazy(() => import("./app/upstream-services/[serviceId]/page"));
/**
 * Summary: UpstreamServicesPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const UpstreamServicesPage = lazy(() => import("./app/upstream-services/page"));
/**
 * Summary: UsersPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const UsersPage = lazy(() => import("./app/users/page"));
/**
 * Summary: VisualizerPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const VisualizerPage = lazy(() => import("./app/visualizer/page"));
/**
 * Summary: WebhooksPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const WebhooksPage = lazy(() => import("./app/webhooks/page"));
/**
 * Summary: UniversalAgentBusPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const UniversalAgentBusPage = lazy(() => import("./app/universal-agent-bus/page"));

/**
 * Summary: PageFallback component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const PageFallback = () => (
  <div className="flex items-center justify-center h-full min-h-[200px]">
    <Loader2 className="animate-spin h-8 w-8 text-muted-foreground" />
  </div>
);

/**
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
