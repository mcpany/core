const fs = require('fs');
const file = 'server/pkg/app/api_webhooks.go';
let content = fs.readFileSync(file, 'utf8');

content = content.replace(
`		case http.MethodPut:
			wh, ok := a.WebhooksManager.GetWebhook(id)
			if !ok {
				http.NotFound(w, r)
				return
			}`,
`		case http.MethodPut:
			_, ok := a.WebhooksManager.GetWebhook(id)
			if !ok {
				http.NotFound(w, r)
				return
			}`
);

fs.writeFileSync(file, content);
