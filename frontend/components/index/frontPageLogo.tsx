import { BodyShort, Heading } from "@navikt/ds-react";

export const FrontPageLogo = () => (
  <div className="flex flex-col items-center">
    <div className="flex items-center md:gap-2 gap-1">
      <div className="nada-slash h-7 w-4 md:h-[2.5rem] md:w-5" />
      <Heading level="1" size="xlarge" className="text-[2rem] md:text-[2.5rem] pb-1 font-bold">datamarkedsplassen</Heading>
      <div className="h-7 w-7 md:h-[2.5rem] md:w-[2.5rem] nada-logo" />
    </div>
    <BodyShort className="font-bold! text-xl! md:text-2xl">dele, finne og bruke data</BodyShort>
  </div>
)

export const HeaderLogo = () => (
  <div className="flex items-center gap-1">
    <div className="nada-slash--white h-4 w-2 hidden md:block" />
    <Heading level="1" size="xsmall" className="text-base font-bold hidden md:block">datamarkedsplassen</Heading>
    <div className="h-4 w-4 nada-logo" />
  </div>

)