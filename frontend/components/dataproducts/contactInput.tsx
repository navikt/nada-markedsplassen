import { TextField } from '@navikt/ds-react'
import { useState } from 'react'
import { FieldValues, FormState, UseFormRegister } from 'react-hook-form'
import { isEmail } from '../../lib/validators'
import { EmailIcon } from '../lib/icons/emailIcon'
import { ErrorIcon } from '../lib/icons/errorIcon'
import { LoaderIcon } from '../lib/icons/loaderIcon'
import { SlackLogo } from '../lib/icons/slackLogo'
import { useIsValidSlackChannel } from '../../lib/rest/slack'

export interface IContactInputProps {
  register: UseFormRegister<any>
  formState: FormState<FieldValues>
}

const InfoLoading = ()=>(
  <div className='flex flex-row items-center'>
    <LoaderIcon />
    <p className='italic text-base'>Sjekker om kontaktpunktet finnes</p>
  </div>
)

const InfoEmail = ()=>(
  <div className='flex flex-row items-center'>
    <EmailIcon />
    <p className='italic text-base'>E-post-adressen blir brukt som kontaktpunkt</p>
  </div>
)

const InfoSlack = ()=>(
  <div className='flex flex-row items-center'>
    <SlackLogo />
    <p className='italic text-base'>Slack-kanalen blir brukt som kontaktpunktet</p>
  </div>
)

const InfoError = ()=>(
  <div className='flex flex-row items-center'>
    <ErrorIcon />
    <p className='italic text-base'>Finner ikke kontaktpunktet. Bruk en e-postadresse eller slack-kanal</p>
  </div>
)

export const ContactInput = ({ register, formState }: IContactInputProps) => {
  const { errors } = formState
  const [input, setInput] = useState('')
  const { isValid, loading } = useIsValidSlackChannel(input)

  return (
    <div>
      <div>
        <TextField
          label="Ønsket kontaktpunkt for dataproduktet"
          description="Kontaktpunktet kan være enten navnet på en public slack-kanal (uten #) eller en e-post."
          {...register('teamContact')}
          error={errors.teamContact?.message?.toString()}
          className="w-full"
          onBlur={(e) => setInput(e.target.value)}
        />
      </div>
      {isEmail(input) && <InfoEmail />}
      {!!input && !isEmail(input) && loading && <InfoLoading />}
      {!loading && isValid && <InfoSlack />}
      {!!input &&
        !loading &&
        !isEmail(input) &&
        !isValid
        && <InfoError />}
    </div>
  )
}
