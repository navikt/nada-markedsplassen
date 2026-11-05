import { CopyButton } from '@navikt/ds-react'

type CopyProps = {
  text: string
}

const Copy = ({ text }: CopyProps) => (
  <CopyButton copyText={text} size="xsmall" />
)

export default Copy
