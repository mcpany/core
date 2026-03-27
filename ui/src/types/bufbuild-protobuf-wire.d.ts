/**
 * Type definitions for the @bufbuild/protobuf/wire module.
 */
declare module '@bufbuild/protobuf/wire' {
  /**
   * Wire types used in protobuf encoding.
   */
  export enum WireType {
    Varint = 0,
    Fixed64 = 1,
    LengthDelimited = 2,
    StartGroup = 3,
    EndGroup = 4,
    Fixed32 = 5,
  }
}
