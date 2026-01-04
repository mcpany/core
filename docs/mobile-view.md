# Mobile View Support

I have implemented a comprehensive "Mobile View" update for the MCP Any Dashboard. This ensures that the application is fully functional and aesthetically pleasing on smaller screens (phones and tablets).

## Features

- **Responsive Layouts:**
    - The **Network Graph** control card now collapses on mobile to save screen space, and the details sheet takes up the full width for better readability.
    - The **Log Stream** header controls (Search, Filter, Pause/Resume) now stack vertically or collapse into icons on mobile devices.
    - The **Secrets Manager** table is now horizontally scrollable, and the "Add Secret" dialog inputs stack vertically for better usability.

- **Mobile UX Enhancements:**
    - Used `useIsMobile` hook to conditionally render mobile-optimized components.
    - Adjusted padding and font sizes where necessary.

## Verification

The feature has been verified using Playwright tests simulating an iPhone viewport.

### Screenshots

**Network Graph (Mobile)**
![Network Graph Mobile](.audit/ui/2026-01-04/mobile_network.png)

**Log Stream (Mobile)**
![Log Stream Mobile](.audit/ui/2026-01-04/mobile_logs.png)

**Secrets Dialog (Mobile)**
![Secrets Dialog Mobile](.audit/ui/2026-01-04/mobile_secrets_dialog.png)
