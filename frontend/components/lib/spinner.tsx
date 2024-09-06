import { Loader } from '@navikt/ds-react'

export const LoaderSpinner = () => (
  <div className="flex flex-grow justify-center items-center">
    <Loader size="2xlarge" transparent />
  </div>
)

export default LoaderSpinner
