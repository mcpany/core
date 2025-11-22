# Reddit

This service allows you to search for submissions in a given subreddit.

## Usage

You can use this service to search for posts in a subreddit.

To search for posts, call the `search` tool with the `subreddit` and `q` parameters:

```
gemini -m gemini-2.5-flash -p 'search for "mcp" in the "golang" subreddit'
```

## Authentication

This service does not require authentication for public subreddits. However, there are rate limits for unauthenticated requests.
