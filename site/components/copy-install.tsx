'use client';

import { useState } from 'react';
import { IconCheck, IconCopy } from '@/components/icons';

export function CopyInstall() {
  const [copied, setCopied] = useState(false);
  const command = 'curl -fsSL https://raw.githubusercontent.com/RavaniRoshan/refweir/main/scripts/install.sh | bash';

  const onCopy = async () => {
    try {
      await navigator.clipboard.writeText(command);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // ignore
    }
  };

  return (
    <div className="flex w-full max-w-2xl items-center justify-between gap-3 rounded-md border border-hairline bg-raised px-4 py-3 font-mono text-xs text-text-1">
      <div className="flex items-center gap-2 overflow-x-auto whitespace-nowrap scrollbar-none">
        <span className="text-text-3 select-none">$</span>
        <span>{command}</span>
      </div>
      <button
        type="button"
        onClick={onCopy}
        aria-label="Copy install command"
        className="flex shrink-0 items-center gap-1.5 rounded-sm border border-hairline bg-slab px-2.5 py-1 text-xs text-text-2 hover:text-text-1 hover:border-patina transition-colors cursor-pointer"
      >
        {copied ? (
          <>
            <IconCheck className="size-3.5 text-patina" />
            <span className="text-patina">Copied</span>
          </>
        ) : (
          <>
            <IconCopy className="size-3.5" />
            <span>Copy</span>
          </>
        )}
      </button>
    </div>
  );
}
