declare module "@bufbuild/protobuf/wire" {
  /**
   * Reads binary data from a buffer.
   *
   * Summary: Provides methods to read binary encoded protocol buffer data.
   *
   * Parameters: None.
   *
   * Returns: None.
   *
   * Throws: None.
   */
  export class BinaryReader {
    constructor(bytes: Uint8Array);
    len: number;
    pos: number;
    [key: string]: any;
  }

  /**
   * Writes binary data to a buffer.
   *
   * Summary: Provides methods to write binary encoded protocol buffer data.
   *
   * Parameters: None.
   *
   * Returns: None.
   *
   * Throws: None.
   */
  export class BinaryWriter {
    constructor();
    [key: string]: any;
  }
}
