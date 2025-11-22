# Stripe

This service allows you to create a charge in Stripe.

## Usage

You can use this service to create a new charge. You will need to provide the amount, currency, and a payment source.

```
gemini -m gemini-2.5-flash -p 'create a new charge for 1000 cents in usd with source tok_visa'
```

## Authentication

This service requires a Stripe API key. You will need to set the `STRIPE_API_KEY` environment variable to your API key. The `config.yaml` file is already configured to use this environment variable.
