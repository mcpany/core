with open('ui/src/components/marketplace/wizard/wizard-context.tsx', 'r') as f:
    content = f.read()

content = content.replace('webhooks: any[]; // TODO: Define webhook type', 'webhooks: { name: string, webhook: { url: string, timeout: string, webhookSecret: string } }[];')

with open('ui/src/components/marketplace/wizard/wizard-context.tsx', 'w') as f:
    f.write(content)
