import * as React from 'react'
import { Copy as CopyIcon } from '@navikt/ds-icons'
import { useState } from 'react'

let timeout: NodeJS.Timeout | undefined = undefined

type CopyProps = {
  text: string
}
const Copy = ({ text }: CopyProps) => {
  const [copied, setCopied] = useState(false)

  const copyToClipboard = (
    e: React.MouseEvent<SVGSVGElement, MouseEvent>,
    id: string | undefined
  ) => {
    id && navigator.clipboard.writeText(id)
    setCopied(true)
    if (timeout) {
      clearTimeout(timeout)
    }
    timeout = setTimeout(() => {
      setCopied(false)
      timeout = undefined
    }, 1500)
  }
  return (
    <div className="inline relative">
      <CopyIcon
        className="text-surface-action cursor-pointer hover:text-surface-action-hover hover:h-4"
        onClick={(e) => copyToClipboard(e, text)}
      />
      {copied && (
        <div className="absolute left-5 -top-2 px-2 p-1 rounded bg-gray-100 border border-gray-300">
          kopiert
        </div>
      )}
    </div>
  )
}

export default Copy
