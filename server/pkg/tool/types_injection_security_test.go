// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

func TestCheckForShellInjection_Security(t *testing.T) {
	tests := []struct {
		name        string
		val         string
		template    string
		placeholder string
		command     string
		shouldFail  bool
	}{
		// Python
		{
			name:        "Python Safe",
			val:         "hello",
			template:    "print('{{input}}')",
			placeholder: "{{input}}",
			command:     "python",
			shouldFail:  false,
		},
		{
			name:        "Python __import__ bypass attempt",
			val:         "'+[c for c in ().__class__.__base__.__subclasses__() if c.__name__ == 'catch_warnings'][0]()._module.__builtins__['__im'+'port__']('sub'+'process').call(['date'])+'",
			template:    "print('{{input}}')",
			placeholder: "{{input}}",
			command:     "python",
			shouldFail:  true, // Should be blocked by __ check
		},
		{
			name:        "Python getattr attempt",
			val:         "'+getattr(os, 'system')('date')+'",
			template:    "print('{{input}}')",
			placeholder: "{{input}}",
			command:     "python",
			shouldFail:  true, // Should be blocked by getattr keyword
		},
		{
			name:        "Python simple __ check",
			val:         "__builtins__",
			template:    "print('{{input}}')",
			placeholder: "{{input}}",
			command:     "python",
			shouldFail:  true,
		},

		// Node
		{
			name:        "Node Safe",
			val:         "hello",
			template:    "console.log('{{input}}')",
			placeholder: "{{input}}",
			command:     "node",
			shouldFail:  false,
		},
		{
			name:        "Node global.process bypass attempt",
			val:         "\\'); global.process.mainModule.constructor._load(\"child\"+\"_process\").execSync(\"date\");//",
			template:    "console.log('{{input}}')",
			placeholder: "{{input}}",
			command:     "node",
			shouldFail:  true, // Should be blocked by global or process keyword
		},
		{
			name:        "Node process.exit",
			val:         "\\'); process.exit();//",
			template:    "console.log('{{input}}')",
			placeholder: "{{input}}",
			command:     "node",
			shouldFail:  true, // Should be blocked by process keyword
		},
        {
            name:        "Node execSync",
            val:         "\\'); require('child_process').execSync('date');//",
            template:    "console.log('{{input}}')",
            placeholder: "{{input}}",
            command:     "node",
            shouldFail:  true, // Should be blocked by execSync or require
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkForShellInjection(tt.val, tt.template, tt.placeholder, tt.command)
			if (err != nil) != tt.shouldFail {
				t.Errorf("checkForShellInjection() error = %v, shouldFail %v", err, tt.shouldFail)
			}
		})
	}
}
