const http = require('http');

const data = JSON.stringify({
    name: 'test-seed-service',
    command_line_service: {
        command: 'echo',
        resources: [
            {
                name: 'Application Logs',
                uri: `mcp://test-seed-service/logs.txt`,
                mime_type: 'text/plain',
                description: 'Logs',
                static: {
                    text_content: 'some application logs here'
                }
            },
            {
                name: 'User Database',
                uri: `mcp://test-seed-service/db.json`,
                mime_type: 'application/json',
                description: 'Database',
                static: {
                    text_content: '{"users": [{"id": 1, "name": "Alice"}]}'
                }
            }
        ]
    }
});

const req = http.request('http://127.0.0.1:35285/api/v1/services', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'Content-Length': data.length
    }
}, (res) => {
    let rawData = '';
    res.on('data', (chunk) => { rawData += chunk; });
    res.on('end', () => {
        console.log("Status:", res.statusCode);
        console.log("Response:", rawData);
    });
});

req.write(data);
req.end();
