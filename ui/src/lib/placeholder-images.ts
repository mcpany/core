/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import data from './placeholder-images.json';

/**
 * ImagePlaceholder represents the public ImagePlaceholder entity.
 *
 * Summary: Defines the structured data model representing a placeholder.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Throws/Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export type ImagePlaceholder = {
  id: string;
  description: string;
  imageUrl: string;
  imageHint: string;
};

/**
 * PlaceHolderImages serves as a public interface for interacting with PlaceHolderImages.
 *
 * Summary: Defines the structured data model representing a holder images.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Throws/Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export const PlaceHolderImages: ImagePlaceholder[] = data.placeholderImages;
