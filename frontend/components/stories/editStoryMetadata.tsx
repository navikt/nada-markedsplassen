import { yupResolver } from '@hookform/resolvers/yup'
import { useForm } from 'react-hook-form'
import ErrorMessage from '../lib/error'
import { useRouter } from 'next/router'
import TeamkatalogenSelector from '../lib/teamkatalogenSelector'
import DescriptionEditor from '../lib/DescriptionEditor'
import {
  Button,
  Heading,
  TextField,
  Select,
} from '@navikt/ds-react'
import amplitudeLog from '../../lib/amplitude'
import * as yup from 'yup'
import { useContext, useState } from 'react'
import TagsSelector from '../lib/tagsSelector'
import {UserState} from "../../lib/context";
import { updateStory } from '../../lib/rest/stories'

const schema = yup.object().shape({
  name: yup.string().nullable().required('Skriv inn navnet på datafortellingen'),
  description: yup.string(),
  teamkatalogenURL: yup.string().required('Du må velge team i teamkatalogen'),
  keywords: yup.array(),
  group: yup.string(),
})


export interface EditStoryMetadataFields {
  id: string
  name: string
  description: string
  keywords: string[]
  teamkatalogenURL: string
  group: string
}

export const EditStoryMetadataForm = ({id, name, description, keywords, teamkatalogenURL, group}: EditStoryMetadataFields) => {
  const router = useRouter()
  const [productAreaID, setProductAreaID] = useState<string>('')
  const [teamID, setTeamID] = useState<string>('')
  const userInfo = useContext(UserState)
  const [error, setError] = useState<Error | undefined>(undefined)
  const [loading, setLoading] = useState(false)
  const {
    register,
    handleSubmit,
    watch,
    formState,
    setValue,
    control,
  } = useForm({
    resolver: yupResolver(schema),
    defaultValues: {
      name: name,
      description: description,
      keywords: keywords,
      teamkatalogenURL: teamkatalogenURL,
      group: group,
    },
  })

  const { errors } = formState
  const kw = watch('keywords')
  
  const onDeleteKeyword = (keyword: string) => {
    kw !== undefined ? 
    setValue('keywords', kw.filter((k: string) => k !== keyword))
    :
    setValue('keywords', [])
  }

  const onAddKeyword = (keyword: string) => {
    kw
      ? setValue('keywords', [...kw, keyword])
      : setValue('keywords', [keyword])
  }

  const onSubmit = async (data: any) => {
    const editStoryData = {
        name: data.name,
        description: data.description,
        keywords: data.keywords,
        teamkatalogenURL: data.teamkatalogenURL,
        productAreaID: productAreaID,
        teamID: teamID,
        group: data.group,
    }

    setLoading(true)
    updateStory(id, editStoryData).then(()=>{
      setError(undefined)
      amplitudeLog('skjema fullført', { skjemanavn: 'endre-datafortelling' })
      router.push("/user/stories")
    }).catch(e=>{
      setError(e)
      console.log(e)
      amplitudeLog('skjemainnsending feilet', {
        skjemanavn: 'endre-datafortelling',
      })
    }).finally(()=>{
      setLoading(false)
    })
  }

  const onCancel = () => {
      amplitudeLog('Klikker på: Avbryt',
      {
        pageName: 'endre-datafortelling',
      })
      router.back()
  }

  const onError = (errors: any) => {
    amplitudeLog('skjemavalidering feilet', {
      skjemanavn: 'endre-datafortelling',
      feilmeldinger: Object.keys(errors)
        .map((errorKey) => errorKey)
        .join(','),
    })
  }

  const gcpProjects = userInfo?.gcpProjects as any[] || []

  return (
    <div className="mt-8 md:w-[46rem]">
      <Heading level="1" size="large">
        Endre datafortelling metadata
      </Heading>
      <form
        className="pt-12 flex flex-col gap-10"
        onSubmit={handleSubmit(onSubmit, onError)}
      >
        <TextField
          className="w-full"
          label="Navn på datafortelling"
          {...register('name')}
          error={errors.name?.message?.toString()}
        />
        <DescriptionEditor
          label="Beskrivelse av hva datafortellingen kan brukes til"
          name="description"
          control={control}
        />
        <Select
          className="w-full"
          label="Velg gruppe fra GCP"
          {...register('group', {
            onChange: () => setValue('teamkatalogenURL', ''),
          })}
          error={errors.group?.message?.toString()}
        >
          <option value="">Velg gruppe</option>
          {[
            ...new Set(
              gcpProjects.map(
                ({ group }: { group: { name: string } }) => (
                  <option
                    value={
                      userInfo?.googleGroups.filter((g: any) => g.name === group.name)[0]
                        .email
                    }
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
          gcpGroups={userInfo?.gcpProjects.map((it: any)=> it.group.email)}
          register={register}
          watch={watch}
          errors={errors}
          setProductAreaID={setProductAreaID}
          setTeamID={setTeamID}
        />
        <TagsSelector
            onAdd={onAddKeyword}
            onDelete={onDeleteKeyword}
            tags={kw || []}
        />
        {error && <ErrorMessage error={error} />}
        <div className="flex flex-row gap-4 mb-16">
          <Button type="button" variant="secondary" onClick={onCancel}>
            Avbryt
          </Button>
          <Button type="submit" disabled={loading}>Lagre</Button>
        </div>
      </form>
    </div>
  )
}
