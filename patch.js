const fs = require('fs');
const file = 'server/pkg/app/api_webhooks.go';
let content = fs.readFileSync(file, 'utf8');

content = content.replace(
`		case http.MethodGet:
			wh, ok := a.WebhooksManager.GetWebhook(id)
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(wh)
		case http.MethodDelete:`,
`		case http.MethodGet:
			wh, ok := a.WebhooksManager.GetWebhook(id)
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(wh)
		case http.MethodPut:
			wh, ok := a.WebhooksManager.GetWebhook(id)
			if !ok {
				http.NotFound(w, r)
				return
			}
			var cfg webhooks.WebhookConfig
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfg.ID = id
			a.WebhooksManager.AddWebhook(&cfg)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(cfg)
		case http.MethodDelete:`
);

fs.writeFileSync(file, content);
