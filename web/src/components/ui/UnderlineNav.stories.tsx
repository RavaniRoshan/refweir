import type { Meta, StoryObj } from '@storybook/react'
import { UnderlineNav } from './UnderlineNav'
import { Code, GitPullRequest, GitCommit, ShieldAlert, Sparkles } from 'lucide-react'

const meta: Meta<typeof UnderlineNav> = {
  title: 'UI/UnderlineNav',
  component: UnderlineNav,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof UnderlineNav>

const sampleTabs = [
  { id: 'code', label: 'Code', icon: <Code className="w-4 h-4" /> },
  { id: 'issues', label: 'Issues', count: 12 },
  { id: 'pulls', label: 'Pull Requests', count: 4, icon: <GitPullRequest className="w-4 h-4" /> },
  { id: 'commits', label: 'Commits', count: '1.4k', icon: <GitCommit className="w-4 h-4" /> },
  { id: 'agents', label: 'Agent Fleet', count: 3, icon: <Sparkles className="w-4 h-4 text-primary" /> },
  { id: 'security', label: 'Security & Policy', icon: <ShieldAlert className="w-4 h-4" /> },
]

export const Default: Story = {
  args: {
    tabs: sampleTabs,
    activeTab: 'code',
  },
}

export const PullRequestsActive: Story = {
  args: {
    tabs: sampleTabs,
    activeTab: 'pulls',
  },
}

export const WithActions: Story = {
  args: {
    tabs: sampleTabs,
    activeTab: 'agents',
    actions: (
      <div className="flex items-center gap-2">
        <span className="text-xs text-muted-foreground font-mono">100% Lease Containment</span>
      </div>
    ),
  },
}
