/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

declare module "@bufbuild/protobuf/wire" {
  export class BinaryReader {
    constructor(bytes: Uint8Array);
    len: number;
    pos: number;
    [key: string]: any;
  }

  export class BinaryWriter {
    constructor();
    [key: string]: any;
  }
}
