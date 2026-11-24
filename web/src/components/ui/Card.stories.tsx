import type { Meta, StoryObj } from '@storybook/react'
import { Box, BoxHeader, BoxBody, BoxRow, BoxFooter } from './Card'
import { Button } from './Button'
import { Badge } from './Badge'
import { Server } from 'lucide-react'

const meta: Meta<typeof Box> = {
  title: 'UI/Card',
  component: Box,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof Box>

export const DefaultCard: Story = {
  render: () => (
    <Box className="max-w-md">
      <BoxHeader className="bg-secondary/35 py-2.5 px-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Server className="w-4 h-4 text-primary" />
          <span className="text-xs font-semibold text-foreground">Cluster Node Telemetry</span>
        </div>
        <Badge variant="success" pulse>Online</Badge>
      </BoxHeader>
      <BoxBody className="p-4 space-y-2 text-xs text-muted-foreground">
        <p>Pure-Go distributed raft cluster node active on 127.0.0.1:8080.</p>
        <div className="font-mono text-foreground text-[11px] bg-secondary/30 p-2 rounded border border-border">
          Storage Engine: modernc.org/sqlite WAL
        </div>
      </BoxBody>
      <BoxFooter className="p-3 border-t border-border flex justify-end gap-2">
        <Button variant="outline" size="sm">Metrics</Button>
        <Button variant="primary" size="sm">Inspect Node</Button>
      </BoxFooter>
    </Box>
  ),
}

export const RowList: Story = {
  render: () => (
    <Box className="max-w-md">
      <BoxHeader className="bg-secondary/35 py-2 px-3 text-xs font-semibold text-foreground">
        Recent Activity
      </BoxHeader>
      <BoxBody className="p-0">
        <BoxRow className="p-3 hover:bg-secondary/20 transition-colors flex items-center justify-between text-xs">
          <span>refweir-agent created PR #104</span>
          <span className="text-muted-foreground text-[11px]">2m ago</span>
        </BoxRow>
        <BoxRow className="p-3 hover:bg-secondary/20 transition-colors flex items-center justify-between text-xs">
          <span>Security scanning passed for commit dff791c</span>
          <span className="text-muted-foreground text-[11px]">8m ago</span>
        </BoxRow>
      </BoxBody>
    </Box>
  ),
}
