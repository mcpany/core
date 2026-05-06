/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import {genkit} from 'genkit';
import {googleAI} from '@genkit-ai/google-genai';

/**
 * Summary: The initialized Genkit instance configured with Google AI plugin and Gemini 2.5 Flash model. Used for AI-powered features in the application.
 *
 * Params:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export const ai = genkit({
  plugins: [googleAI()],
  model: 'googleai/gemini-2.5-flash',
});
