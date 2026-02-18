const http = require('http');
const url = require('url');

const PORT = 9999;

const server = http.createServer((req, res) => {
  const parsedUrl = url.parse(req.url, true);

  // Set CORS headers
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');

  if (req.method === 'OPTIONS') {
    res.writeHead(200);
    res.end();
    return;
  }

  console.log(`[OAuth Server] ${req.method} ${req.url}`);

  if (parsedUrl.pathname === '/auth') {
    // Expecting query params: response_type=code, client_id, redirect_uri, state
    const redirectUri = parsedUrl.query.redirect_uri;
    const state = parsedUrl.query.state;

    if (!redirectUri) {
      res.writeHead(400);
      res.end('Missing redirect_uri');
      return;
    }

    // Immediately redirect back with a code
    const code = 'mock-auth-code';
    const callbackUrl = `${redirectUri}?code=${code}&state=${state || ''}`;

    console.log(`[OAuth Server] Redirecting to ${callbackUrl}`);
    res.writeHead(302, { Location: callbackUrl });
    res.end();
  } else if (parsedUrl.pathname === '/token') {
    // Expecting POST with grant_type=authorization_code, code, redirect_uri, client_id
    // Just return a dummy token
    const tokenResponse = {
      access_token: 'mock-access-token',
      token_type: 'Bearer',
      expires_in: 3600,
      refresh_token: 'mock-refresh-token',
      scope: 'read write'
    };

    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(tokenResponse));
  } else if (parsedUrl.pathname === '/userinfo') {
    const userInfo = {
      sub: 'mock-user-id',
      name: 'Mock User',
      email: 'mock@example.com'
    };
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(userInfo));
  } else {
    res.writeHead(404);
    res.end('Not Found');
  }
});

server.listen(PORT, () => {
  console.log(`Fake OAuth Server listening on port ${PORT}`);
});
