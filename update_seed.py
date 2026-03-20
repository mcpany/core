import re

with open('/app/ui/tests/e2e/test-data.ts', 'r') as f:
    content = f.read()

replacement = """                    { name: "process_payment", description: "Process a payment", call_id: "process_payment_call", input_schema: { type: "object", properties: { amount: { type: "number", description: "Payment amount in cents" }, currency: { type: "string", description: "Currency code (e.g., USD)" } }, required: ["amount", "currency"] } }"""

content = re.sub(
    r'\{\s*name:\s*"process_payment",\s*description:\s*"Process a payment",\s*call_id:\s*"process_payment_call"\s*\}',
    replacement,
    content
)

with open('/app/ui/tests/e2e/test-data.ts', 'w') as f:
    f.write(content)
