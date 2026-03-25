/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

declare module "@bufbuild/protobuf/wire" {
  export class BinaryReader {
    constructor(bytes: Uint8Array);
    len: number;
    pos: number;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    [key: string]: any;
  }

  export class BinaryWriter {
    constructor();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    [key: string]: any;
  }
}
