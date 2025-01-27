import { ChevronDownIcon, ChevronUpIcon } from '@navikt/aksel-icons'
import { Button } from '@navikt/ds-react'
import { ReactNode, useState } from 'react'

export interface FoldablePanelProps {
  caption: string
  children?: ReactNode
  className?: string
}

export const FoldablePanel = ({
  className,
  caption,
  children,
}: FoldablePanelProps) => {
  const [open, setOpen] = useState(false)
  return (
    <div className={className}>
      <Button variant="tertiary" onClick={() => setOpen(!open)} className="panel-item">
        {caption}
        {open ? <ChevronUpIcon title="a11y-title" fontSize="1.5rem" />
          : <ChevronDownIcon title="a11y-title" fontSize="1.5rem" />}
      </Button>
      <div
        className={`
          transition-property:max-height duration-200 ease-in-out
          ${open ? 'mt-2 max-h-[999rem]' : 'max-h-0 invisible overflow-y-hidden'}`}
      >
        {children}
      </div>
    </div>
  )
}
