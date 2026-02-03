# MCP Any UI

The **MCP Any UI** provides a modern, interactive dashboard for managing your MCP Any server. It allows you to visualize the network topology, monitor metrics (latency, QPS), and manage services, tools, and resources.

## 🚀 Getting Started

This is a [Next.js](https://nextjs.org/) project bootstrapped with [`create-next-app`](https://github.com/vercel/next.js/tree/canary/packages/create-next-app).

### Prerequisites

- **Node.js**: 18.17 or later.
- **MCP Any Server**: A running instance of the `mcpany` server.

### Installation

1.  Navigate to the `ui` directory:
    ```bash
    cd ui
    ```

2.  Install dependencies:
    ```bash
    npm install
    # or
    yarn install
    # or
    pnpm install
    ```

### Running Locally

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:9002](http://localhost:9002) with your browser to see the result.

You can start editing the page by modifying `src/app/page.tsx`. The page auto-updates as you edit the file.

### Testing

Run the end-to-end tests with Playwright:

```bash
npx playwright test
```

## 📂 Project Structure

- **`src/app`**: Next.js App Router pages and layouts.
- **`src/components`**: Reusable UI components (shadcn/ui, custom components).
- **`src/lib`**: Utility functions and client SDKs.
- **`docs`**: Documentation for UI features.

## ✨ Features

- **Network Topology**: Visualize your MCP ecosystem.
- **Service Management**: Add, edit, and monitor upstream services.
- **[Interactive Playground](docs/playground.md)**: Test tools with auto-generated forms supporting complex schemas.
- **Observability**: Real-time metrics and logs.

## 📚 Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.
