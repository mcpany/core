const fs = require('fs');
const file = 'ui/src/components/users/user-list.tsx';
let content = fs.readFileSync(file, 'utf8');

content = content.replace(
`                                            {user.authentication?.apiKey || (user.authentication as any)?.api_key ? (
                                                <div className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-muted/50 border">
                                                    <Key className="h-3.5 w-3.5 text-orange-500" />
                                                    <span>API Key</span>
                                                </div>
                                            ) : user.authentication?.basicAuth || (user.authentication as any)?.basic_auth ? (`,
`                                            {user.authentication?.apiKey || (user.authentication as any)?.api_key ? (
                                                <div className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-muted/50 border">
                                                    <Key className="h-3.5 w-3.5 text-orange-500" />
                                                    <span>API Key</span>
                                                </div>
                                            ) : user.authentication?.basicAuth || (user.authentication as any)?.basic_auth ? (`
);

fs.writeFileSync(file, content);
