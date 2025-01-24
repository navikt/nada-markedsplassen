import { XMarkOctagonFillIcon } from "@navikt/aksel-icons"

interface errorMessageProps {
  error: Error
}

export const ErrorMessage = ({ error }: errorMessageProps) => {
  return (
    <div className="bg-surface-danger-subtle rounded px-1 py-2 h-fit">
      <div className="flex items-center text-base gap-1">
        <XMarkOctagonFillIcon title="a11y-title" fontSize="1.5rem" color="#C30000" />
        Feil
      </div>
      <p>{error.message}</p>
    </div>
  )
}

export default ErrorMessage
