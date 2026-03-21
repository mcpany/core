import json

with open('ui/tsconfig.json', 'r') as f:
    data = json.load(f)

if 'paths' not in data['compilerOptions']:
    data['compilerOptions']['paths'] = {}

data['compilerOptions']['paths']["../../../proto/*"] = ["../proto/*"]
data['compilerOptions']['paths']["@proto/*"] = ["../proto/*"]
data['compilerOptions']['paths']["@bufbuild/protobuf/wire"] = ["./node_modules/@bufbuild/protobuf/dist/esm/wire/index.js"]
data['compilerOptions']['paths']["@improbable-eng/grpc-web"] = ["./node_modules/@improbable-eng/grpc-web"]
data['compilerOptions']['paths']["browser-headers"] = ["./node_modules/browser-headers"]
data['compilerOptions']['paths']["long"] = ["./node_modules/long"]

with open('ui/tsconfig.json', 'w') as f:
    json.dump(data, f, indent=2)
