# Stripe

This service allows you to create a charge in Stripe.

## Usage

You can use this service to create a new charge. You will need to provide the amount, currency, and a payment source.

```
gemini -m gemini-2.5-flash -p 'create a new charge for 1000 cents in usd with source tok_visa'
```

## Authentication

This service requires a Stripe API key. You will need to set the `STRIPE_API_KEY` environment variable to your API key. The `config.yaml` file is already configured to use this environment variable.

## Usage with Gemini CLI

### 1. Start the MCP Server

```bash
# From repo root
make build # if not already built
# Export required environment variables
export STRIPE_API_KEY=YOUR_STRIPE_API_KEY_VALUE

./build/bin/server run --config-path examples/popular_services/stripe/config.yaml
```

### 2. Add to Gemini

In a separate terminal:

```bash
gemini mcp add --transport http --trust stripe http://localhost:50050
```

### 3. Example Query

```bash
gemini -m gemini-2.5-flash -p "List the available tools for this service"
```
