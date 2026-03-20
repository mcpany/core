const fs = require('fs');
const filepath = 'server/pkg/app/api_traces.go';
let content = fs.readFileSync(filepath, 'utf8');

// I modified `Result` to be an array. Let's see if the type of `Result` in `audit.Entry` allows it.
