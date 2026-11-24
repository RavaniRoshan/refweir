import type { Meta, StoryObj } from '@storybook/react'
import { DocsView } from './DocsView'

const meta: Meta<typeof DocsView> = {
  title: 'Views/DocsView',
  component: DocsView,
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof DocsView>

export const QuickstartArticle: Story = {
  args: {
    initialArticleId: 'quickstart',
  },
}

export const CapabilityLeasesArticle: Story = {
  args: {
    initialArticleId: 'capability-leases',
  },
}

export const McpToolsSpecification: Story = {
  args: {
    initialArticleId: 'mcp-tools-reference',
  },
}

export const RulesetsPolicyArticle: Story = {
  args: {
    initialArticleId: 'rulesets-policy',
  },
}
