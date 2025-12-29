import '@testing-library/jest-dom';

// ResizeObserver mock
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// PointerEvents mock (often needed for radix-ui)
if (typeof window.PointerEvent === 'undefined') {
  class PointerEvent extends MouseEvent {
    pointerId: number;
    pointerType: string;
    isPrimary: boolean;

    constructor(type: string, params: PointerEventInit = {}) {
      super(type, params);
      this.pointerId = params.pointerId || 0;
      this.pointerType = params.pointerType || '';
      this.isPrimary = params.isPrimary || false;
    }
  }
  (window as any).PointerEvent = PointerEvent;
}
