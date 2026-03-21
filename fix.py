with open('ui/src/components/register-service-dialog.tsx', 'r') as f:
    content = f.read()

# Let's fix line 634 which we added `const res: any`
# It violates the @typescript-eslint/no-explicit-any rule in CI perhaps (even though it's a warning locally)
import re
content = content.replace("const res: any = await apiClient.registerService(configToSave);", "const res = await apiClient.registerService(configToSave) as { service?: { id?: string; sanitizedName?: string; name?: string } };")

# I don't see any other `catch (e: any)` remaining but let's be sure.
with open('ui/src/components/register-service-dialog.tsx', 'w') as f:
    f.write(content)
