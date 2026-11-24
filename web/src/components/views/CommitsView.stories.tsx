import type { Meta, StoryObj } from '@storybook/react'
import { CommitsView } from './CommitsView'

const meta: Meta<typeof CommitsView> = {
  title: 'Views/CommitsView',
  component: CommitsView,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof CommitsView>

export const DefaultTimeline: Story = {
  args: {
    repo: {
      ownerSlug: 'acme',
      name: 'rocket',
      defaultBranch: 'main',
      cloneUrlHttp: 'http://localhost:8080/acme/rocket.git',
      cloneUrlSsh: 'ssh://git@localhost:2222/acme/rocket.git',
    },
  },
}
