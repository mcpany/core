# Twilio

This service provides communication APIs for SMS, voice, and more.

## Usage

You can use this service to send SMS messages.

To send an SMS, provide the `to`, `from`, and `body` parameters:

```
gemini -m gemini-2.5-flash -p 'send an SMS to +1234567890 from +10987654321 with the message "Hello from Twilio!"'
```

## Authentication

This service requires an Account SID and Auth Token. Set the `TWILIO_ACCOUNT_SID` and `TWILIO_AUTH_TOKEN` environment variables.
