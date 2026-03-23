import re

filepath = 'server/pkg/app/server_init.go'
with open(filepath, 'r') as f:
    content = f.read()

# Remove the previously injected code
pattern = r'\s*// Add swarm mock topology nodes to seed the database[\s\S]*?if err := storeStorage\.SaveService\(ctx, s\); err != nil \{\s*log\.Error\("Failed to seed swarm service", "error", err\)\s*\}\s*\}'
content = re.sub(pattern, '', content)

# Find where weatherService is saved, and inject the swarm nodes there
injection = """
	// Add swarm mock topology nodes to seed the database
	swarmOrchestrator := configv1.UpstreamServiceConfig_builder{
		Id:   proto.String("n1"),
		Name: proto.String("Primary Orchestrator"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: proto.String("echo"),
		}.Build(),
		Tags: []string{"validator"},
	}.Build()

	researchAgent := configv1.UpstreamServiceConfig_builder{
		Id:   proto.String("n2"),
		Name: proto.String("Research Agent"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: proto.String("echo"),
		}.Build(),
		Tags: []string{"agent"},
	}.Build()

	toolExec := configv1.UpstreamServiceConfig_builder{
		Id:   proto.String("n3"),
		Name: proto.String("Tool Exec"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: proto.String("echo"),
		}.Build(),
		Tags: []string{"service"},
	}.Build()

	synthesizer := configv1.UpstreamServiceConfig_builder{
		Id:   proto.String("n4"),
		Name: proto.String("Synthesizer"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: proto.String("echo"),
		}.Build(),
		Tags: []string{"agent"},
	}.Build()

	rogueNode := configv1.UpstreamServiceConfig_builder{
		Id:   proto.String("n5"),
		Name: proto.String("Rogue Node"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: proto.String("exit 1"),
		}.Build(),
		Tags: []string{"agent"},
	}.Build()

	for _, s := range []*configv1.UpstreamServiceConfig{swarmOrchestrator, researchAgent, toolExec, synthesizer, rogueNode} {
		if err := storeStorage.SaveService(ctx, s); err != nil {
			log.Error("Failed to seed swarm service", "error", err)
		}
	}
"""

pattern = r'(if err := storeStorage\.SaveService\(ctx, weatherService\); err != nil \{\s*log\.Error\("Failed to save default weather service", "error", err\)\s*return err\s*\})'
content = re.sub(pattern, r'\1\n' + injection, content)

with open(filepath, 'w') as f:
    f.write(content)
