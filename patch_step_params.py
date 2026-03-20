with open('ui/src/components/marketplace/wizard/steps/step-parameters.tsx', 'r') as f:
    content = f.read()

def replace_sync(content):
    # This robustly syncs the params mapping to the correct config path depending on the service type defined
    old_sync_logic = """        // Also update config env
        // TODO: Sync `params` to `config.commandLineService.env` more robustly
        // For now we just update basic env
        if (config.commandLineService) {
            const env: any = {};
            Object.entries(newParams).forEach(([k, v]) => {
                env[k] = { plainText: v };
            });
            updateConfig({
                commandLineService: {
                    ...config.commandLineService,
                    env
                }
            });
        }"""

    new_sync_logic = """        // Also update config env
        const env: any = {};
        Object.entries(newParams).filter(([k, v]) => k.trim() !== '').forEach(([k, v]) => {
            env[k] = { plainText: v };
        });

        if (config.commandLineService) {
            updateConfig({
                commandLineService: {
                    ...config.commandLineService,
                    env
                }
            });
        } else if (config.mcpService?.stdioConnection) {
             updateConfig({
                 mcpService: {
                     ...config.mcpService,
                     stdioConnection: {
                         ...config.mcpService.stdioConnection,
                         env
                     }
                 }
             });
        }"""

    old_remove_sync = """         // Sync with config
         if (config.commandLineService) {
            const env: any = {};
            Object.entries(newParams).forEach(([k, v]) => {
                env[k] = { plainText: v };
            });
            updateConfig({
                commandLineService: {
                    ...config.commandLineService,
                    env
                }
            });
        }"""

    new_remove_sync = """         // Sync with config
         const env: any = {};
         Object.entries(newParams).filter(([k, v]) => k.trim() !== '').forEach(([k, v]) => {
             env[k] = { plainText: v };
         });

         if (config.commandLineService) {
             updateConfig({
                 commandLineService: {
                     ...config.commandLineService,
                     env
                 }
             });
         } else if (config.mcpService?.stdioConnection) {
             updateConfig({
                 mcpService: {
                     ...config.mcpService,
                     stdioConnection: {
                         ...config.mcpService.stdioConnection,
                         env
                     }
                 }
             });
         }"""

    content = content.replace(old_sync_logic, new_sync_logic)
    content = content.replace(old_remove_sync, new_remove_sync)
    return content

content = replace_sync(content)

with open('ui/src/components/marketplace/wizard/steps/step-parameters.tsx', 'w') as f:
    f.write(content)
