const fs = require('fs');
const file = 'ui/src/components/users/user-sheet.tsx';
let content = fs.readFileSync(file, 'utf8');

content = content.replace(
`                    id: user.id,
                    role: user.roles[0] || "viewer",
                    authType: user.authentication?.apiKey || (user.authentication as any)?.api_key ? "api_key" : "password",`,
`                    id: user.id,
                    role: (user.roles[0] as "admin" | "editor" | "viewer") || "viewer",
                    authType: user.authentication?.apiKey || (user.authentication as any)?.api_key ? "api_key" : "password",`
);

fs.writeFileSync(file, content);
