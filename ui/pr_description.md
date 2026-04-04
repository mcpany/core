> **🎯 Strategic Focus:** Fix Log View Broken Window
>
> **🔍 Origin Story:**
> * **Source:** Path B: UX Gap
> * **The Pain:** When the user checks the logs containing JSON arrays, the UI renders them as raw text dumps rather than leveraging the rich `SmartTable` component. This decreases readability, usability, and goes against the enterprise-grade polished product vision.
>
> **🛠️ The Solution:**
> * **Feature:** Array JSON payload mapping in LogViewer
> * **Design:** Update the `getTableData` function in `json-view.tsx` to handle primitive data types appropriately inside of top-level or nested array structures by transforming arrays into `{ index, value }` dictionaries, ensuring arrays are converted into standard table rows when presented in the UI.
>
> **🏗️ Architecture:**
> * **Data Strategy:** Log entries emitted from tracing are naturally seeded into the actual backend. The UI interprets real payload returns by parsing JSON arrays within the logs and feeding them to the generic `SmartTable` pipeline, removing any need for localized network mocks.
> * **Testing:** `tests/logs-smart-table.spec.ts` verifies that the Playwright execution traces generate an array JSON dump. Upon expansion, the logs evaluate and instantiate the `SmartTable` container properly via realistic trace fetching.
>
> **✅ Verification:**
> * [x] Local Tests Passed
> * [x] E2E Suite Passed (Real Data)
> * [x] Linting Passed