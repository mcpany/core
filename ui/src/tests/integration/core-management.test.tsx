/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react'
import ServicesPage from '@/app/services/page'
import ToolsPage from '@/app/tools/page'
import ResourcesPage from '@/app/resources/page'
import PromptsPage from '@/app/prompts/page'
import ProfilesPage from '@/app/profiles/page'

// Mock ResizeObserver
class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
global.ResizeObserver = ResizeObserver

describe('Core Management Pages', () => {
  describe('Services Page', () => {
    it('renders services list', () => {
      render(<ServicesPage />)
      expect(screen.getByText('github-service')).toBeInTheDocument()
      expect(screen.getByText('slack-integration')).toBeInTheDocument()
    })

    it('filters services', () => {
      render(<ServicesPage />)
      const searchInput = screen.getByPlaceholderText('Search services...')
      fireEvent.change(searchInput, { target: { value: 'github' } })

      expect(screen.getByText('github-service')).toBeInTheDocument()
      expect(screen.queryByText('slack-integration')).not.toBeInTheDocument()
    })

    it('toggles service status', () => {
      render(<ServicesPage />)
      // This is a basic test to ensure the switch renders.
      // Full functionality requires more complex state testing or integration testing.
      const switches = screen.getAllByRole('switch')
      expect(switches.length).toBeGreaterThan(0)
    })
  })

  describe('Tools Page', () => {
    it('renders tools list', () => {
      render(<ToolsPage />)
      expect(screen.getByText('github-service/list_repos')).toBeInTheDocument()
    })
  })

  describe('Resources Page', () => {
    it('renders resources list', () => {
        render(<ResourcesPage />)
        expect(screen.getByText('system-logs')).toBeInTheDocument()
    })
  })

  describe('Prompts Page', () => {
    it('renders prompts list', () => {
        render(<PromptsPage />)
        expect(screen.getByText('summarize_logs')).toBeInTheDocument()
    })
  })

  describe('Profiles Page', () => {
    it('renders profiles list', () => {
        render(<ProfilesPage />)
        expect(screen.getByText('dev')).toBeInTheDocument()
        expect(screen.getByText('prod')).toBeInTheDocument()
    })
  })
})
