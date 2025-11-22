# Google Translate

This service provides language translation for a given text.

## Usage

You can use this service to translate text from one language to another.

To translate text, provide the `text`, `source_language`, and `target_language` parameters:

```
gemini -m gemini-2.5-flash -p 'translate "Hello, world!" from English to Spanish'
```

## Authentication

This service requires an API key. Set the `GOOGLE_API_KEY` environment variable to your API key.
