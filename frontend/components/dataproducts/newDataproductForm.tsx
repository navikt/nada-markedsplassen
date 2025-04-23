import { yupResolver } from '@hookform/resolvers/yup'
import {
  Button,
  Heading,
  Select,
  TextField,
} from '@navikt/ds-react'
import { useRouter } from 'next/router'
import { useContext, useState } from 'react'
import { useForm } from 'react-hook-form'
import * as yup from 'yup'
import { UserState } from '../../lib/context'
import { createDataproduct } from '../../lib/rest/dataproducts'
import DescriptionEditor from '../lib/DescriptionEditor'
import ErrorStripe from "../lib/errorStripe"
import TeamkatalogenSelector from '../lib/teamkatalogenSelector'
import { ContactInput } from './contactInput'


const schema = yup.object().shape({
  name: yup.string().nullable().required('Du må fylle inn navn'),
  description: yup.string(),
  team: yup
    .string()
    .required('Velg en gruppe fra GCP som skal ha ansvar for dataproduktet'),
  teamkatalogenURL: yup.string().required('Du må velge et team i Teamkatalogen'),
  teamContact: yup.string().nullable(),
})

export const NewDataproductForm = () => {
  const router = useRouter()
  const userInfo = useContext(UserState)
  const [productAreaID, setProductAreaID] = useState<string>('')
  const [teamID, setTeamID] = useState<string>('')
  const [backendError, setBackendError] = useState<Error|undefined>(undefined)

  const {
    register,
    handleSubmit,
    watch,
    formState,
    setValue,
    getValues,
    control,
    trigger
  } = useForm({
    resolver: yupResolver(schema),
    defaultValues: {  
        name: undefined,
        description: '',
        team: '',
        teamContact: null,
        teamkatalogenURL: undefined,
    }
  })

  const { errors } = formState
  const dataproductName = watch('name')
  const description = watch('description')
  const team = watch('team')
  const teamkatalogenURL = watch('teamkatalogenURL')
  const teamContact = watch('teamContact')

  const submitForm = async () => {
      createDataproduct({
            name: dataproductName,
            group: team,
            description: description,
            teamkatalogenURL: teamkatalogenURL,
            teamContact: teamContact?? undefined,
            productAreaID: productAreaID || undefined,
            teamID: teamID || undefined,
      }).then((res) => {
        const data = res
        setBackendError(undefined)
        router.push(`/dataproduct/${data.id}/${data.slug}`)
    }).catch (e=> {
      setBackendError(e)
      console.log(e)
    })
  }

  const onCancel = () => {
    router.back()
  }

  const gcpProjects = userInfo?.gcpProjects as any[] || []
  return (
    <div className="mt-8 md:w-[46rem]">
      <Heading level="1" size="large">
        Legg til dataprodukt
      </Heading>
      <form
        className="pt-12 flex flex-col gap-10"
        onSubmit={handleSubmit(submitForm)}
      >
        <TextField
          className="w-full"
          label="Navn på dataprodukt"
          {...register('name')}
          error={errors.name?.message?.toString()}
        />
        <DescriptionEditor
          label="Beskrivelse av hva dataproduktet kan brukes til"
          name="description"
          control={control}
        />
        <Select
          className="w-full"
          label="Velg gruppe fra GCP"
          {...register('team', {
            onChange: () => setValue('teamkatalogenURL', ''),
          })}
          error={errors.team?.message?.toString()}
        >
          <option value="">Velg gruppe</option>
          {[
            ...new Set(
              gcpProjects.map(
                ({ group }: { group: { name: string, email: string } }) => (
                  <option
                    value={group.email}
                    key={group.name}
                  >
                    {group.name}
                  </option>
                )
              )
            ),
          ]}
        </Select>
        <TeamkatalogenSelector
          gcpGroups={[team]}
          register={register}
          watch={watch}
          errors={errors}
          setProductAreaID={setProductAreaID}
          setTeamID={setTeamID}
        />
        <ContactInput register={register} formState={formState} />
        {backendError && <ErrorStripe error={backendError} />}
        <div className="flex flex-row gap-4 mb-16">
          <Button type="button" variant="secondary" onClick={onCancel}>
            Avbryt
          </Button>
          <div className="flex flex-row gap-4">
            <Button
                variant="primary"
                size="medium"
            >
                Lagre dataprodukt
            </Button>
          </div>
        </div>
      </form>
    </div>
  )
}
