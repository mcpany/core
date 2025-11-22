# Twilio

This service allows you to send an SMS message using Twilio.

## Usage

You can use this service to send an SMS message to a phone number. You will need to provide the `To` and `From` phone numbers, and the `Body` of the message.

```
gemini -m gemini-2.5-flash -p 'send a message to +15551234567 from +15557654321 with body "Hello, world!"'
```

## Authentication

This service requires a Twilio Account SID and Auth Token. You will need to set the `TWILIO_ACCOUNT_SID` and `TWILIO_AUTH_TOKEN` environment variables. The `config.yaml` file is already configured to use these environment variables.
