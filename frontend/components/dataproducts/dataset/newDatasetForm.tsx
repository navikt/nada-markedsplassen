import { yupResolver } from '@hookform/resolvers/yup'
import { Button, Checkbox, Heading, Loader, Radio, RadioGroup, Textarea, TextField } from '@navikt/ds-react'
import { useRouter } from 'next/router'
import { Controller,useForm } from 'react-hook-form'
import * as yup from 'yup'
import DescriptionEditor from '../../lib/DescriptionEditor'
import TagsSelector from '../../lib/tagsSelector'
import DatasetSourceForm from './datasetSourceForm'
import { useColumnTags } from './useColumnTags'
import { TilgangsstyringHelpText } from "./helptext";
import { PiiForm } from './piiForm'
import { useState } from 'react'
import { createDataset } from '../../../lib/rest/dataproducts'
import ErrorStripe from "../../lib/errorStripe";

/** PiiLevel defines all possible levels of personal identifiable information that a dataset can have. */
export enum PiiLevel {
  Anonymised = 'anonymised',
  None = 'none',
  Sensitive = 'sensitive'
}

interface NewDatasetFormProps {
  dataproduct: any
}

export type FormValues = {
  name: string,
  description?: string | undefined,
  repo?: string | null | undefined,
  bigquery: {
    dataset: string,
    projectID: string,
    table: string,
  }
  pii: NonNullable<PiiLevel | null | undefined>,
  keywords?: any[] | undefined,
  anonymisationDescription?: string | null | undefined,
  grantAllUsers?: string | null | undefined,
  teamInternalUse?: boolean | undefined,
}

const schema = yup.object().shape({
  name: yup.string().nullable().required('Du må fylle inn navn'),
  description: yup.string(),
  repo: yup.string().nullable(),
  bigquery: yup.object({
    dataset: yup.string().required(),
    projectID: yup.string().required(),
    table: yup.string().required(),
  }),
  pii: yup
    .mixed<PiiLevel>()
    .nullable()
    .oneOf([PiiLevel.Sensitive, PiiLevel.Anonymised, PiiLevel.None])
    .required(
      'Du må spesifisere om datasettet inneholder personidentifiserende informasjon'
    ),
  keywords: yup.array(),
  anonymisationDescription: yup.string().nullable().when("pii", {
    is: "anonymised",
    then: () => yup.string().nullable().required('Du må beskrive hvordan datasettet har blitt anonymisert')
  }),
  grantAllUsers: yup.string().nullable(),
  teamInternalUse: yup.boolean(),
})

const NewDatasetForm = ({ dataproduct }: NewDatasetFormProps) => {
  const router = useRouter()
  const {
    register,
    handleSubmit,
    control,
    watch,
    setValue,
    getValues,
    formState,
  } = useForm({
    resolver: yupResolver(schema),
    defaultValues: {
      name: '',
      description: '',
      repo: '',
      bigquery: { projectID: '', dataset: '', table: '' },
      pii: 'none' as PiiLevel,
      keywords: [] as string[],
      anonymisationDescription: '',
      grantAllUsers: 'dontGrantAllUsers',
      teamInternalUse: undefined,
    } as FormValues,
  })
  const [submitted, setSubmitted] = useState(false)
  const errors = formState.errors
  const projectID = watch('bigquery.projectID')
  const datasetID = watch('bigquery.dataset')
  const tableID = watch('bigquery.table')
  const [backendError, setBackendError] = useState<any>(undefined)
  const pii = watch('pii')
  const {
    columns,
    loading: loadingColumns,
    error: columnsError,
    tags,
    pseudoColumns,
    annotateColumn,
    selectPseudoColumn,
  } = useColumnTags(projectID, datasetID, tableID)

  const onDeleteKeyword = (keyword: string) => {
    setValue(
      'keywords',
      keywords !== undefined ? keywords.filter((k: string) => k !== keyword) : keywords
    )
  }

  const keywords = watch('keywords')
  const team = dataproduct.owner.group

  const onAddKeyword = (keyword: string) => {
    keywords
      ? setValue('keywords', [...keywords, keyword])
      : setValue('keywords', [keyword])
  }

  const onSubmitForm = async (requestData: any) => {
    setSubmitted(true)
    requestData.dataproductID = dataproduct.id
    requestData.bigquery.piiTags = JSON.stringify(Object.fromEntries(tags || new Map<string, string>()))
    const pii = requestData.pii === "sensitive"
      ? PiiLevel.Sensitive
      : requestData.pii === "anonymised"
        ? PiiLevel.Anonymised
        : PiiLevel.None
    requestData.pii = pii
    requestData.grantAllUsers = requestData.pii === PiiLevel.Sensitive || requestData.grantAllUsers === '' ? null : requestData.grantAllUsers === 'grantAllUsers'
    requestData.targetUser = requestData.teamInternalUse ? "OwnerTeam" : ""
    requestData.teamInternalUse = undefined
    requestData.pseudoColumns = pseudoColumns ? Array.from(pseudoColumns.entries())
      .filter(([, value]) => value)
      .map(([key]) => key) : []
    try {
      const data = await createDataset(requestData)
      router.push(
        `/dataproduct/${dataproduct.id}/${dataproduct.slug}/${data.id}`
      )
      setBackendError(undefined)
    } catch (e) {
      setBackendError(e)
    }
  }
  const selectedAllColumns = Array.from(pseudoColumns).filter(e => e[1]).length === columns?.length

  return (
    <div className="pt-8 pr-8 md:w-[46rem]">
      <Heading level="1" size="medium" spacing>
        Legg til datasett
      </Heading>
      <Checkbox {...register('teamInternalUse')}>Datasettet er ment til bruk innad i teamet</Checkbox>

      <form
        onSubmit={handleSubmit(onSubmitForm)}
        className="flex flex-col gap-10 h-[90%]"
      >
        <TextField
          {...register('name')}
          className="w-full"
          label="Skriv inn navn"
          error={errors?.name?.message?.toString()}
          size="medium"
        />
        <DescriptionEditor
          {...register('description')}
          name="description"
          label="Beskrivelse av hva datasettet kan brukes til"
          control={control}
        />
        <TextField
          {...register('repo')}
          type={'url'}
          className="w-full"
          id="repo"
          label="Link til kildekode"
          error={errors.repo?.message?.toString()}
        />
        <DatasetSourceForm
          label="Velg tabell eller view"
          team={team}
          register={register}
          errors={errors}
          setValue={setValue}
        />
        <TagsSelector
          onAdd={onAddKeyword}
          onDelete={onDeleteKeyword}
          tags={keywords || []}
        />
        <PiiForm loading={false}
          error={undefined}
          columns={columns}
          tags={tags}
          pseudoColumns={pseudoColumns}
          control={control}
          getValues={getValues}
          register={register}
          formState={formState}
          watch={watch}
          annotateColumn={annotateColumn}
          pseudoynimiseColumn={selectPseudoColumn}
        />
        <Controller name="grantAllUsers" control={control} render={({ field }) => (
          <RadioGroup {...field} legend={<p className="flex gap-2 items-center">
            Tilgangsstyring
            <TilgangsstyringHelpText />
          </p>
          }>
            <Radio value="dontGrantAllUsers">Brukere må søke om tilgang til datasettet</Radio>
            <Radio value="grantAllUsers" disabled={![PiiLevel.None, PiiLevel.Anonymised].includes(getValues('pii'))}>
              Gi tilgang til alle i Nav {![PiiLevel.None, PiiLevel.Anonymised].includes(getValues('pii')) && "(kan ikke gi tilgang til alle i Nav når datasettet inneholder personopplysninger)"}
            </Radio>
          </RadioGroup>
        )}
        />
        {backendError && <ErrorStripe error={backendError} />}
        {submitted && !backendError && <div>Vennligst vent...<Loader size="small" /></div>}
        <div className="flex flex-row gap-4 grow items-end">
          <Button
            type="button"
            variant="secondary"
            onClick={() => {
              router.push(
                `/dataproduct/${dataproduct.dataproduct.id}/${dataproduct.dataproduct.slug}/info`
              )
            }}
          >
            Avbryt
          </Button>
          <Button type="submit" disabled={false && (selectedAllColumns || submitted)}>Lagre</Button>
        </div>
      </form>
    </div>
  )
}

export default NewDatasetForm
