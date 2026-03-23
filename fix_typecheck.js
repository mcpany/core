const fs = require('fs');

const path = 'ui/src/components/dashboard/download-report-button.tsx';
if (fs.existsSync(path)) {
    let code = fs.readFileSync(path, 'utf-8');
    // We only want to replace (s) => with (s: any) =>
    code = code.replace(/\(s\) =>/g, '(s: any) =>');
    fs.writeFileSync(path, code);
    console.log("Fixed " + path);
} else {
    console.log("File " + path + " not found!");
}
