import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

console.log(fs.existsSync(path.join(__dirname, "../bazel-out/k8-fastbuild/bin/proto")));
