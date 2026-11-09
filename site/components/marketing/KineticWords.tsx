'use client';

import { useEffect, useState } from 'react';

const WORDS = ['coding agents', 'AI agents', 'your agents'];

export function KineticWords() {
  const [index, setIndex] = useState(0);
  const [reduced, setReduced] = useState(false);

  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)');
    setReduced(mq.matches);
    if (mq.matches) return;
    const id = window.setInterval(() => {
      setIndex((i) => (i + 1) % WORDS.length);
    }, 2600);
    return () => window.clearInterval(id);
  }, []);

  if (reduced) {
    return <span className="marker text-patina">coding&nbsp;agents.</span>;
  }

  return (
    <span className="relative inline-block align-baseline">
      {WORDS.map((word, i) => (
        <span
          key={word}
          aria-hidden={i !== index}
          className={`marker text-patina absolute left-0 top-0 whitespace-nowrap transition-all duration-500 ease-out ${
            i === index
              ? 'opacity-100 translate-y-0'
              : 'opacity-0 translate-y-2 pointer-events-none'
          }`}
        >
          {word}.
        </span>
      ))}
      {/* Spacer to reserve width for the longest word */}
      <span className="marker text-patina invisible whitespace-nowrap" aria-hidden="true">
        coding&nbsp;agents.
      </span>
    </span>
  );
}
