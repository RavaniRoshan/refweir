import type { Meta, StoryObj } from '@storybook/react'
import { MergeBox } from './MergeBox'

const meta: Meta<typeof MergeBox> = {
  title: 'UI/MergeBox',
  component: MergeBox,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof MergeBox>

export const ReadyToMerge: Story = {
  args: {
    checksPassed: true,
    defaultMethod: 'squash',
    defaultTitle: 'feat(core): implement pure-Go raft state machine',
    defaultMessage: 'Squashed commits from agent/raft-cluster',
    onMerge: (method, title) => alert(`Merged with ${method}: ${title}`),
  },
}

export const BlockedByReview: Story = {
  args: {
    checksPassed: true,
    agentReviewBlocker: {
      blocked: true,
      summary: 'Policy ruleset violation: secret detected in commit history (high entropy string in .env.example)',
      violationsCount: 1,
    },
    onMerge: () => {},
  },
}

export const AlreadyMerged: Story = {
  args: {
    isMerged: true,
    onMerge: () => {},
  },
}
