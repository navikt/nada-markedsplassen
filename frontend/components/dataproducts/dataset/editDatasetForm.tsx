import {
  Button,
  Checkbox,
  ErrorSummary,
  Heading,
  Radio,
  RadioGroup,
  Textarea,
  TextField,
} from '@navikt/ds-react'
import { Controller, useForm } from 'react-hook-form'
import { yupResolver } from '@hookform/resolvers/yup';
import { useState } from 'react'
import * as yup from 'yup'
import DescriptionEditor from '../../lib/DescriptionEditor'
import TagsSelector from '../../lib/tagsSelector';
import { useColumnTags } from './useColumnTags';
import AnnotateDatasetTable from './annotateDatasetTable';
import {Personopplysninger} from "./helptext";
import { useRouter } from 'next/router';
import { updateDataset } from '../../../lib/rest/dataproducts';
import { DatasetWithAccess } from '../../../lib/rest/generatedDto';

interface EditDatasetFormProps {
  dataset: DatasetWithAccess
  setEdit: (value: boolean) => void
}

interface EditDatasetFormFields {
  name: string
  description?: string | undefined
  repo?: string | null | undefined
  pii: string
  bigquery: {
    projectID: string
    dataset: string
    table: string
  }
  keywords?: any[] | undefined
  anonymisationDescription?: string | undefined
  teamInternalUse?: boolean
}

const schema = yup.object().shape({
  name: yup.string().nullable().required('Du må fylle inn navn'),
  description: yup.string(),
  repo: yup.string().nullable(),
  pii: yup.string().nullable()
    .oneOf(['anonymised','none','sensitive'])
    .required(
      'Du må spesifisere om datasettet inneholder personidentifiserende informasjon'
    ),
  bigquery: yup.object({
    dataset: yup.string().required(),
    projectID: yup.string().required(),
    table: yup.string().required(),
  }),
  keywords: yup.array(),
  anonymisationDescription: yup.string().when("pii", {
    is: "anonymised",
    then: () => yup.string().nullable().required('Du må beskrive hvordan datasettet har blitt anonymisert')
  }),
  teamInternalUse: yup.boolean(),
})

const EditDatasetForm = ({ dataset, setEdit }: EditDatasetFormProps) => {
  const [backendError, setBackendError] = useState<Error>()
  const router = useRouter()

  const { register, handleSubmit, watch, formState, setValue, getValues, control } =
    useForm({
      resolver: yupResolver(schema),
      defaultValues: {
        name: dataset?.name,
        description: dataset?.description || '',
        pii: dataset?.pii as ("anonymised" | "none" | "sensitive"),
        repo: dataset?.repo || '',
        keywords: dataset?.keywords,
        bigquery: {
          projectID: dataset?.datasource?.projectID,
          dataset: dataset?.datasource?.dataset,
          table: dataset?.datasource?.table,
        },
        anonymisationDescription: dataset?.anonymisationDescription ? dataset.anonymisationDescription : '',
        teamInternalUse: dataset?.targetUser === "OwnerTeam",
      },
    })

  const keywords = watch('keywords')
  const bigquery = watch('bigquery')
  const pii = watch('pii')
  const {
    columns,
    loading: loadingColumns,
    error: columnsError,
    tags,
    pseudoColumns,
    annotateColumn,
    selectPseudoColumn,
  } = useColumnTags(bigquery.projectID, bigquery.dataset, bigquery.table, dataset)

  const onDeleteKeyword = (keyword: string) => {
    setValue(
      'keywords',
      keywords !== undefined ? keywords.filter((k: string) => k !== keyword) : [] 
    )
  }

  const onAddKeyword = (keyword: string) => {
    keywords !== undefined ? setValue('keywords', [...keywords, keyword]) : setValue('keywords', [keyword])
  }

  const { errors } = formState
  const onSubmit = (requestData: any) => {
    console.log("debug on submit edit dataset", requestData)
    const payload= {
      name: requestData.name,
      description: requestData.description,
      pii: requestData.pii,
      repo: requestData.repo?? undefined,
      keywords: requestData.keywords??[],
      anonymisationDescription: requestData.anonymisationDescription,
      targetUser: requestData.teamInternalUse? "OwnerTeam" : "",
      piiTags: JSON.stringify(Object.fromEntries(tags || new Map<string, string>())),
      pseudoColumns: Array.from(pseudoColumns).filter(it=> it[1]).map(it=> it[0]),
    }
    updateDataset(dataset.id, payload ).then(() => {
      setBackendError(undefined)
      setEdit(false)
      router.reload()
    }).catch((e:Error) => {
      setBackendError(e)
    })
  }
  const hasPseudoColumns = !!dataset.datasource?.pseudoColumns?.length
  const selectedAllColumns = Array.from(pseudoColumns).filter(e=> e[1]).length === columns?.length
  return (
    <div className="block pt-8 pr-8 md:w-[46rem]">
      <Heading level="1" size="medium" spacing>
        Endre datasett
      </Heading>
      <Checkbox {...register('teamInternalUse')}>Datasettet er ment til bruk innad i teamet</Checkbox>
      <form
        onSubmit={handleSubmit(onSubmit)}
        className="flex flex-col gap-10 h-[90%]"
      >
        <TextField
          {...register('name')}
          label="Skriv inn navn"
          error={errors?.name?.message}
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
          id="repo"
          label="Link til kildekode"
          error={errors.repo?.message}
        />
        <TextField
          label="Dataset"
          disabled
          value={
            bigquery.projectID + '.' + bigquery.dataset + '.' + bigquery.table
          }
        />
        <TagsSelector
          onAdd={onAddKeyword}
          onDelete={onDeleteKeyword}
          tags={keywords || []}
        />
        <Controller
          name="pii"
          control={control}
          render={({ field }) => (
            <RadioGroup
              {...field}
              legend={<p className="flex gap-2 items-center">Inneholder datasettet personopplysninger? <Personopplysninger /></p>}
              error={errors?.pii?.message}
            >
              {dataset.pii === 'sensitive' && <Radio value={"sensitive"}>
                Ja, inneholder personopplysninger
              </Radio>}
              {pii === 'sensitive' && bigquery.projectID && bigquery.dataset && bigquery.table && (
                <AnnotateDatasetTable
                  loading={loadingColumns}
                  error={columnsError}
                  columns={columns}
                  tags={tags}
                  pseudoColumns={pseudoColumns}
                  annotateColumn={annotateColumn}
                  selectPseudoColumn={hasPseudoColumns? selectPseudoColumn: undefined}
                />
              )}
              {dataset.pii === 'anonymised' && <Radio value={"anonymised"}>
                Det er benyttet metoder for å anonymisere personopplysningene
              </Radio>}
              <Textarea 
                placeholder="Beskriv kort hvordan opplysningene er anonymisert" 
                label="Metodebeskrivelse" 
                aria-hidden={getValues("pii") !== "anonymised"}
                className={getValues("pii") !== "anonymised" ? "hidden" : ""}
                error={errors?.anonymisationDescription?.message}
                {...register("anonymisationDescription")}
              />
              {dataset.pii === 'none' && <Radio value={"none"}>
                Nei, inneholder ikke personopplysninger
              </Radio>}
            </RadioGroup>
          )}
        />
        <div className="flex flex-row gap-4 grow items-end">
          <Button
            type="button"
            variant="secondary"
            onClick={() => {
              setEdit(false)
            }}
          >
            Avbryt
          </Button>
          <Button type="submit" disabled={selectedAllColumns}>Lagre</Button>
        </div>
        {
    backendError && (
      <ErrorSummary heading={`Cannot update dataset: ${backendError.message || backendError.toString()}`}>error</ErrorSummary>
    )
  }

      </form>
    </div>
  )
}
export default EditDatasetForm
