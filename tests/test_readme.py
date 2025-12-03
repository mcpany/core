# Copyright 2025 Author(s) of MCP Any
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# SPDX-License-Identifier: Apache-2.0

import os

def test_readme_for_mcp_any_cli():
    """
    This test checks the README.md file for any occurrences of the string "mcp-any-cli".
    """
    readme_path = os.path.join(os.path.dirname(__file__), '..', 'README.md')
    with open(readme_path, 'r') as f:
        content = f.read()
    assert "mcp-any-cli" not in content, "The string 'mcp-any-cli' should not be present in the README.md file."
