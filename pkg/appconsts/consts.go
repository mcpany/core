/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package appconsts

const (
	// Name is the name of the MCP Any server. This is used in help messages and
	// other user-facing output.
	Name = "mcpany"
)

// Version is the version of the MCP Any server. This is a variable so it can be
// set at build time using ldflags. The default value is "dev", which is used
// for local development builds.
var Version = "dev"
