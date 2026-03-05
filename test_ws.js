const WebSocket = require('ws');
const ws = new WebSocket('ws://localhost:3000/api/v1/ws/traces');

ws.on('open', function open() {
  console.log('connected');
  ws.send(Date.now());
  ws.close();
});

ws.on('error', function error(err) {
  console.error('error:', err);
});
