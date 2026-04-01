> **🎯 Strategic Focus:** Enhance UX for Trace Details & Service Loading

> **🔍 Origin Story:**
> * **Source:** Path B: UX Gap
> * **The Pain:** The `ServicesTable` component lacked a proper loading state, falling back to raw text. More importantly, the `TraceDetail` component in the traces payload view dumped raw, unformatted JSON instead of utilizing the `RichResultViewer`, making the inspection of JSON arrays and large outputs hard to read.

> **🛠️ The Solution:**
> * **Feature:** Replaced simple text with a skeleton loading state in the `ServicesTable`. Upgraded the payload views in `TraceDetail` to use the `RichResultViewer` instead of `JsonView`.
> * **Design:** A polished Unifi/Apple aesthetic is used via the `RichResultViewer` that supports smart table detection, falling back to formatted JSON. The table skeleton utilizes subtle pulsed borders and matching table-layout styling to feel completely seamless.

> **🏗️ Architecture:**
> * **Data Strategy:** I verified that the traces load real data by interacting with the backend and explicitly triggering trace generation (using the `/api/v1/debug/seed` debug endpoint for testing) rather than mocking it.
> * **Testing:** An E2E test `TestDashboardUX_TraceDetailsPayload` was implemented to seed data to the backend before checking interactions, fulfilling the "Real Data Law".

> **✅ Verification:**
> * [x] Local Tests Passed
> * [x] E2E Suite Passed (Real Data)
> * [x] Linting Passed
