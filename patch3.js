const fs = require('fs');
const file = 'ui/src/app/settings/webhooks/page.tsx';
let content = fs.readFileSync(file, 'utf8');

content = content.replace(
`    const toggleWebhook = async (id: string) => {
        // Toggle active status not implemented in backend yet, just placeholder for UI interaction
        // In real impl, this would be a PATCH or PUT
        toast.info("Toggle active status not yet implemented in backend");
    };`,
`    const toggleWebhook = async (id: string) => {
        const hook = webhooks.find(h => h.id === id);
        if (!hook) return;

        try {
            const res = await fetch(\`/api/v1/webhooks/\${id}\`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ ...hook, active: !hook.active })
            });
            if (res.ok) {
                toast.success(\`Webhook \${hook.active ? 'disabled' : 'enabled'}\`);
                fetchWebhooks();
            } else {
                toast.error("Failed to update webhook");
            }
        } catch (error) {
            toast.error("Failed to update webhook");
        }
    };`
);

fs.writeFileSync(file, content);
