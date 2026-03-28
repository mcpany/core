async function test() {
  const response = await fetch('http://localhost:50050/api/v1/services', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': 'test-token'
    },
    body: JSON.stringify({
        name: "test-resource-viewer",
        command_line_service: {
          command: 'echo',
          resources: [
            { uri: 'test://data.json', name: 'JSON Data', mimeType: 'application/json' },
            { uri: 'test://invalid.json', name: 'Invalid JSON', mimeType: 'application/json' }
          ],
          reads: {
            'test://data.json': {
              contents: [
                {
                  uri: 'test://data.json',
                  mimeType: 'application/json',
                  text: JSON.stringify([
                    { name: 'Alice', role: 'Admin', id: 1 },
                    { name: 'Bob', role: 'User', id: 2 }
                  ])
                }
              ]
            }
          }
        }
    })
  });
  console.log(response.status, await response.text());
}
test();
