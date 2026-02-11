#!/usr/bin/env node
/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


const readline = require('readline');

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

rl.on('line', (line) => {
  try {
    const request = JSON.parse(line);
    const id = request.id;

    if (request.method === 'initialize') {
      console.log(JSON.stringify({
        jsonrpc: '2.0',
        id: id,
        result: {
          protocolVersion: '2024-11-05',
          capabilities: {
            tools: {}
          },
          serverInfo: {
            name: 'mock-server',
            version: '1.0.0'
          }
        }
      }));
    } else if (request.method === 'notifications/initialized') {
      // No response needed
    } else if (request.method === 'tools/list') {
      console.log(JSON.stringify({
        jsonrpc: '2.0',
        id: id,
        result: {
          tools: [
            {
              name: 'e2e_echo',
              description: 'E2E Echo Tool',
              inputSchema: {
                type: 'object',
                properties: {
                  message: { type: 'string' }
                },
                required: ['message']
              }
            }
          ]
        }
      }));
    } else if (request.method === 'tools/call') {
      const args = request.params.arguments;
      console.log(JSON.stringify({
        jsonrpc: '2.0',
        id: id,
        result: {
          content: [
            {
              type: 'text',
              text: `Echo: ${args.message}`
            }
          ],
          isError: false
        }
      }));
    } else {
        // Ignore other methods
    }
  } catch (e) {
    // Ignore parse errors
  }
});
