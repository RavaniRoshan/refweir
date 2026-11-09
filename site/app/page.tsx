import { Nav } from '@/components/marketing/Nav';
import { Hero } from '@/components/marketing/Hero';
import { SwarmBand } from '@/components/marketing/SwarmBand';
import { Loop } from '@/components/marketing/Loop';
import { ContainmentDemo } from '@/components/marketing/ContainmentDemo';
import { Product } from '@/components/marketing/Product';
import { Platform } from '@/components/marketing/Platform';
import { Quickstart } from '@/components/marketing/Quickstart';
import { Evidence } from '@/components/marketing/Evidence';
import { FinalCta } from '@/components/marketing/FinalCta';
import { StatusFooter } from '@/components/marketing/StatusFooter';

export default function Home() {
  return (
    <>
      <Nav />
      <main>
        <Hero />
        <SwarmBand />
        <Loop />
        <Product />
        <ContainmentDemo />
        <Platform />
        <Quickstart />
        <Evidence />
        <FinalCta />
      </main>
      <StatusFooter />
    </>
  );
}
