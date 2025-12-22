import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock ResizeObserver
class ResizeObserver {
  observe = vi.fn()
  unobserve = vi.fn()
  disconnect = vi.fn()
}
window.ResizeObserver = ResizeObserver

// Mock pointer capture methods
window.Element.prototype.setPointerCapture = vi.fn()
window.Element.prototype.releasePointerCapture = vi.fn()
window.Element.prototype.hasPointerCapture = vi.fn()
window.Element.prototype.scrollIntoView = vi.fn()
