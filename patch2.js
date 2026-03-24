const fs = require('fs');

const file = 'ui/src/components/users/user-list.tsx';
let content = fs.readFileSync(file, 'utf-8');


content = content.replace(/itemContent=\{\(index, user\) => \(/g, 'itemContent={(index, user) => ( <React.Fragment key={user.id}>');
content = content.replace(/<TableRow key=\{user\.id\} data-testid=\{`user-row-\$\{user\.id\}`\}>/g, '<TableRow data-testid={`user-row-${user.id}`}>');
content = content.replace(/<DropdownMenuItem onClick=\{\(\) => copyToClipboard\(user\.id\)\}>/g, `<DropdownMenuItem onClick={() => copyToClipboard(user.id)}>`);

content = content.replace(/<\/DropdownMenu>\n                                    <\/TableCell>\n                                <\/TableRow>\n                            \)/g, `</DropdownMenu>\n                                    </TableCell>\n                                </TableRow>\n                            </React.Fragment>\n                            )`);


fs.writeFileSync(file, content);
