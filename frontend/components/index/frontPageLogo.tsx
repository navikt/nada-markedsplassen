import { BodyShort, Heading } from "@navikt/ds-react";

export const FrontPageLogo = () => (
  <div className="flex flex-col items-center">
    <div className="flex items-center ax-md:gap-2 gap-1">
      <div className="nada-slash h-7 w-4 ax-md:h-[2.5rem] ax-md:w-5" />
      <Heading level="1" size="xlarge" className="text-[2rem] ax-md:text-[2.5rem] pb-1 font-ax-bold">datamarkedsplassen</Heading>
      <div className="h-7 w-7 ax-md:h-[2.5rem] ax-md:w-[2.5rem] nada-logo" />
    </div>
    <BodyShort className="font-ax-bold text-xl ax-md:text-2xl">dele, finne og bruke data</BodyShort>
  </div>
)

export const HeaderLogo = () => (
  <div className="flex items-center gap-1">
    <div className="nada-slash--white h-4 w-2 hidden ax-md:block" />
    <Heading level="1" size="xsmall" className="text-base font-ax-bold hidden ax-md:block">datamarkedsplassen</Heading>
    <div className="h-4 w-4 nada-logo" />
  </div>

)