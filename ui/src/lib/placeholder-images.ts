/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import data from './placeholder-images.json';

/**
 * Intent: Document ImagePlaceholder
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * ImagePlaceholder type definition.
 */
export type ImagePlaceholder = {
  id: string;
  description: string;
  imageUrl: string;
  imageHint: string;
};

/**
 * Intent: Document PlaceHolderImages
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * The PlaceHolderImages const.
 */
export const PlaceHolderImages: ImagePlaceholder[] = data.placeholderImages;
