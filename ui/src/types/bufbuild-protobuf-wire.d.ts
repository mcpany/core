declare module "@bufbuild/protobuf/wire" {
    /**
   * Summary: Reads binary encoded protocol buffer data.
   *
   * Parameters:
   *   - bytes: Uint8Array. The binary data to read.
   *
   * Returns:
   *   - BinaryReader: The initialized binary reader instance.
   *
   * Throws/Errors:
   *   - Throws error if bytes are invalid or corrupted.
   */
  export class BinaryReader {
    constructor(bytes: Uint8Array);
    len: number;
    pos: number;
    [key: string]: any;
  }

    /**
   * Summary: Writes binary encoded protocol buffer data.
   *
   * Parameters: None.
   *
   * Returns:
   *   - BinaryWriter: The initialized binary writer instance.
   *
   * Throws/Errors: None.
   */
  export class BinaryWriter {
    constructor();
    [key: string]: any;
  }
}
