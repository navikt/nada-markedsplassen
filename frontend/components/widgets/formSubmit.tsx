import { Button, Loader } from '@navikt/ds-react'

interface RightJustifiedSubmitButtonProps {
  onCancel?: () => void
  loading?: boolean
}

export const RightJustifiedSubmitButton = ({
  onCancel,
  loading,
}: RightJustifiedSubmitButtonProps) => (
  <div className="flex w-full justify-end gap-4">
    {onCancel && (
      <Button type={'button'} variant={'danger'} onClick={onCancel}>
        Avbryt
      </Button>
    )}
    <Button type={'submit'}>
      {loading ? (
        <>
          <Loader /> Lagrer
        </>
      ) : (
        'Lagre'
      )}
    </Button>
  </div>
)

export default RightJustifiedSubmitButton
