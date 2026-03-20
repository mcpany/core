import re

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # manual replacements
    if filepath == "server/pkg/tool/hooks.go":
        content = re.sub(r'(?<!// Summary: Represents a ResponseData\.\n)\t*type ResponseData struct \{', r'\t// Summary: Represents a ResponseData.\n\ttype ResponseData struct {', content)

    if filepath == "server/tests/framework/e2e.go":
        content = re.sub(r'(\tFileRegistration RegistrationMethod = "file")', r'\t// Summary: Defines FileRegistration.\n\tFileRegistration RegistrationMethod = "file"', content)
        content = re.sub(r'(\tGRPCRegistration RegistrationMethod = "grpc")', r'\t// Summary: Defines GRPCRegistration.\n\tGRPCRegistration RegistrationMethod = "grpc"', content)
        content = re.sub(r'(\tJSONRPCRegistration RegistrationMethod = "jsonrpc")', r'\t// Summary: Defines JSONRPCRegistration.\n\tJSONRPCRegistration RegistrationMethod = "jsonrpc"', content)

    with open(filepath, 'w') as f:
        f.write(content)

for filepath in ["server/pkg/tool/hooks.go", "server/tests/framework/e2e.go"]:
    process_file(filepath)
