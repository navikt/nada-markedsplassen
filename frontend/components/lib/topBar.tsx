import * as React from 'react'
import { Heading } from '@navikt/ds-react'


interface TopBarProps {
  children?: React.ReactNode
  name: string
  type: string | undefined
}

const TopBar = ({ name, type, children }: TopBarProps) => {
  return (
    <div className="flex flex-col flex-wrap text-text-default py-4 md:px-4 gap-2 border-b border-border-on-inverted">
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
