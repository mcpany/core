// mcpany-seed.js
const http = require('http');

async function seed() {
    const payload = {
        name: "policy-test-service",
        id: "policy-test-service",
        version: "1.0.0",
        priority: 0,
        disable: false,
        command_line_service: {
            command: "echo 'Policy Test'",
            working_directory: "/tmp",
            env: {},
            communication_protocol: 0
        },
        call_policies: [
            {
                default_action: 0, // ALLOW
                rules: [
                    {
                        action: 1, // DENY
                        name_regex: "^delete_.*",
                        argument_regex: ".*DROP TABLE.*"
                    },
                    {
                        action: 1, // DENY
                        name_regex: "^admin_.*",
                        argument_regex: ""
                    }
                ]
            }
        ],
        tags: ["test", "policy"]
    };

    const data = JSON.stringify(payload);

    const options = {
        hostname: 'localhost',
        port: 50050, // Default MCP Any port
        path: '/api/v1/services',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Content-Length': data.length
        }
    };

    const req = http.request(options, (res) => {
        console.log(`STATUS: ${res.statusCode}`);
        res.setEncoding('utf8');
        res.on('data', (chunk) => {
            console.log(`BODY: ${chunk}`);
        });
        res.on('end', () => {
            console.log('No more data in response.');
        });
    });

    req.on('error', (e) => {
        console.error(`problem with request: ${e.message}`);
    });

    req.write(data);
    req.end();
}

seed();
