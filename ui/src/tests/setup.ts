/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Mock PointerEvent
if (!global.PointerEvent) {
  class PointerEvent extends MouseEvent {
    public height: number;
    public isPrimary: boolean;
    public pointerId: number;
    public pointerType: string;
    public pressure: number;
    public tangentialPressure: number;
    public tiltX: number;
    public tiltY: number;
    public twist: number;
    public width: number;

    constructor(type: string, params: PointerEventInit = {}) {
      super(type, params);
      this.height = params.height || 0;
      this.isPrimary = params.isPrimary || false;
      this.pointerId = params.pointerId || 0;
      this.pointerType = params.pointerType || "mouse";
      this.pressure = params.pressure || 0;
      this.tangentialPressure = params.tangentialPressure || 0;
      this.tiltX = params.tiltX || 0;
      this.tiltY = params.tiltY || 0;
      this.twist = params.twist || 0;
      this.width = params.width || 0;
    }
  }
  (global as any).PointerEvent = PointerEvent;
}

// Mock scrollIntoView
window.HTMLElement.prototype.scrollIntoView = vi.fn();
window.HTMLElement.prototype.releasePointerCapture = vi.fn();
window.HTMLElement.prototype.hasPointerCapture = vi.fn();
