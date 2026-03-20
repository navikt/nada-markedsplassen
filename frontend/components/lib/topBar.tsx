import { Heading } from '@navikt/ds-react'
import * as React from 'react'


interface TopBarProps {
  children?: React.ReactNode
  name: string
}

const TopBar = ({ name, children }: TopBarProps) => {
  return (
    <div className="flex flex-col flex-wrap text-ax-text-neutral py-4 md:px-4 gap-2 border-b border-ax-border-neutral-subtle">
      <span className="flex gap-5 items-center">
      {
        // have to ignore in order to use dangerouslySetInnerHTML :(
        // also, <wbr> might not be supported on every mobile browser, but i haven't tested this yet
        //@ts-ignore
        <Heading level="1" size="xlarge" dangerouslySetInnerHTML={{ __html: name.replaceAll("_", "_<wbr>") }}></Heading>
      }
      </span>
      {children}
    </div>
  )
}

export default TopBar
