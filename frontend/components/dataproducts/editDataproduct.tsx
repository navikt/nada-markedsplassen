import { Button, ErrorSummary, Heading, TextField } from '@navikt/ds-react'
import { useForm } from 'react-hook-form'
import { updateDataproductValidation } from '../../lib/schema/yupValidations'
import { yupResolver } from '@hookform/resolvers/yup';
import { useState } from 'react'
import TeamkatalogenSelector from '../lib/teamkatalogenSelector'
import DescriptionEditor from '../lib/DescriptionEditor'
import { useRouter } from 'next/router'
import { ContactInput } from './contactInput';
import { updateDataproduct } from '../../lib/rest/dataproducts';

interface EditDatacollectionFormProps {
  product: any
}

const EditDataproduct = ({ product }: EditDatacollectionFormProps) => {
  const [backendError, setBackendError] = useState()
  const router = useRouter()
  const { register, handleSubmit, watch, formState, setValue, control } =
    useForm({
      resolver: yupResolver(updateDataproductValidation),
      defaultValues: {
        name: product.name,
        description: product.description || '',
        teamkatalogenURL: product.owner.teamkatalogenURL === null ? undefined : product.owner.teamkatalogenURL,
        teamContact: product.owner.teamContact,
        productAreaID: product.owner.productAreaID,
      },
    })
  const [productAreaID, setProductAreaID] = useState<string>('')
  const [teamID, setTeamID] = useState<string>('')

  const { errors } = formState
  const onSubmit = (requestData: any) => {
    requestData.productAreaID = productAreaID
    requestData.teamID = teamID
    updateDataproduct(product.id, requestData).then(() => {
      setBackendError(undefined)
      router.push(`/dataproduct/${product.id}/${product.slug}`)
    }).catch((error) => {
      setBackendError(error)
    })
  }
  return (
    <div className="md:w-[46rem] md:px-4 py-4">
      <Heading level="2" size="large" spacing>
        Endre dataprodukt
      </Heading>
      <form className="flex flex-col gap-10" onSubmit={handleSubmit(onSubmit)}>
        <TextField
          className='w-full'
          style={{ display: 'block' }}
          id="name"
          label="Navn"
          {...register('name')}
          error={errors.name?.message}
        />
        <DescriptionEditor
          label="Beskrivelse"
          name="description"
          control={control}
        />
        <TeamkatalogenSelector
          gcpGroups={[product.owner.group]}
          register={register}
          watch={watch}
          errors={errors}
          setProductAreaID={setProductAreaID}
          setTeamID={setTeamID}
        />
        <ContactInput register={register} formState={formState} />
        <div className="flex flex-row gap-4 grow items-end">
          <Button
            type="button"
            variant="secondary"
            onClick={() => {
              router.push(`/dataproduct/${product.id}/${product.slug}`)
            }}
          >
            Avbryt
          </Button>
          <Button type="submit">Lagre</Button>
          {
            backendError && (
              <ErrorSummary heading={'Feil fra server'}>{backendError}</ErrorSummary>
            )
          }

        </div>
      </form>
    </div>
  )
}
export default EditDataproduct
