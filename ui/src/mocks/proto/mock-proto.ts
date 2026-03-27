/**
 * Mock implementation of the protobuf-wire module for testing.
 */
export const mockProto = {
  /**
   * Encodes a message to wire format.
   * @param message The message to encode.
   * @returns The encoded message.
   */
  encode: (message: any) => message,
  /**
   * Decodes a message from wire format.
   * @param buffer The buffer to decode.
   * @returns The decoded message.
   */
  decode: (buffer: any) => buffer,
};
