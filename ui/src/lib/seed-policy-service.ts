import { apiClient, UpstreamServiceConfig } from "./client";

export const seedCallPolicyService = async () => {
    const serviceConfig: UpstreamServiceConfig = {
        name: "policy-test-service",
        id: "policy-test-service",
        version: "1.0.0",
        priority: 0,
        disable: false,
        commandLineService: {
            command: "echo 'Policy Test'",
            workingDirectory: "/tmp",
            env: {},
            communicationProtocol: 0,
            local: false
        },
        callPolicies: [
            {
                defaultAction: 0, // ALLOW
                rules: [
                    {
                        action: 1, // DENY
                        nameRegex: "^delete_.*",
                        argumentRegex: ".*DROP TABLE.*"
                    },
                    {
                        action: 1, // DENY
                        nameRegex: "^admin_.*",
                        argumentRegex: ""
                    }
                ]
            }
        ],
        tags: ["test", "policy"]
    };

    try {
        await apiClient.registerService(serviceConfig);
        console.log("Service seeded successfully");
    } catch (e) {
        console.error("Failed to seed service:", e);
    }
};

// Execute if run directly
if (require.main === module) {
    seedCallPolicyService();
}
