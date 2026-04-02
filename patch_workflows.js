const fs = require('fs');

const path = '.github/workflows/ci.yml';
let content = fs.readFileSync(path, 'utf8');

// Replace the checkout step with one that cleans up space first
const oldCheckout = `- uses: actions/checkout@v5`;

const newCheckout = `- uses: actions/checkout@v5
      - name: Free disk space
        run: |
          sudo rm -rf /usr/share/dotnet
          sudo rm -rf /opt/ghc
          sudo rm -rf /usr/local/.ghcup
          sudo rm -rf /usr/local/lib/android
          sudo rm -rf /usr/share/swift`;

content = content.replace(oldCheckout, newCheckout);
content = content.replace(oldCheckout, newCheckout); // replace both instances

fs.writeFileSync(path, content, 'utf8');
console.log('Patched workflows successfully');
