/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react'
import MiddlewarePage from '@/app/middleware/page'
import WebhooksPage from '@/app/webhooks/page'

// Mock ReactFlow
vi.mock('@xyflow/react', () => ({
  ReactFlow: ({ children }: any) => <div data-testid="react-flow">{children}</div>,
  Background: () => <div>Background</div>,
  Controls: () => <div>Controls</div>,
  MiniMap: () => <div>MiniMap</div>,
  useNodesState: (initial: any) => [initial, vi.fn(), vi.fn()],
  useEdgesState: (initial: any) => [initial, vi.fn(), vi.fn()],
  addEdge: vi.fn(),
}))

// Mock ResizeObserver
class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
global.ResizeObserver = ResizeObserver

describe('Advanced Features Pages', () => {
  describe('Middleware Page', () => {
    it('renders middleware pipeline container', () => {
      render(<MiddlewarePage />)
      expect(screen.getByText('Middleware Pipeline')).toBeInTheDocument()
      expect(screen.getByTestId('react-flow')).toBeInTheDocument()
    })
  })

  describe('Webhooks Page', () => {
    it('renders webhooks list', () => {
      render(<WebhooksPage />)
      expect(screen.getByText('Webhooks')).toBeInTheDocument()
      expect(screen.getByText('https://api.example.com/hooks/pre-call')).toBeInTheDocument()
    })
  })
})
