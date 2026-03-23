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
