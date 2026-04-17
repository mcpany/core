declare module "@bufbuild/protobuf/wire" {
  /**
   * Summary: Declaration for BinaryReader.
   *
   * Parameters: None
   *
   * Returns: None
   *
   * Errors: None
   */
  export class BinaryReader {
    constructor(bytes: Uint8Array);
    len: number;
    pos: number;
    [key: string]: any;
  }

  /**
   * Summary: Declaration for BinaryWriter.
   *
   * Parameters: None
   *
   * Returns: None
   *
   * Errors: None
   */
  export class BinaryWriter {
    constructor();
    [key: string]: any;
  }
}
