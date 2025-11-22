# Stripe

This service provides payment processing capabilities.

## Usage

You can use this service to create charges, manage customers, and more.

To create a charge, provide the `amount`, `currency`, and `source` parameters:

```
gemini -m gemini-2.5-flash -p 'create a charge of 1000 cents in USD with a test token'
```

## Authentication

This service requires an API key. Set the `STRIPE_API_KEY` environment variable to your API key.
