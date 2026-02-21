# Mobile View Feature

I have implemented mobile responsiveness improvements for the MCP Any dashboard.

## Features

1. **Network Graph Mobile Mode**
   - The control card is collapsible to save screen space.
   - Controls are repositioned for accessibility on smaller screens.
   - The node details sheet takes full width on mobile devices.

2. **Log Stream Mobile Layout**
   - Header controls stack vertically to fit narrow screens.
   - Log rows are adjusted for readability, hiding less critical columns on small viewports.

3. **Secret Manager Mobile Layout**
   - The list of secrets is responsive and scrollable.
   - The "Add Secret" dialog inputs are stacked for better usability.

## Implementation Verification

The mobile responsiveness is achieved through:
- **Tailwind CSS breakpoints** (e.g., `md:hidden`, `flex-col`) in React components.
- **Conditional rendering** based on viewport size (e.g., `isMobile` checks in Network Graph).
