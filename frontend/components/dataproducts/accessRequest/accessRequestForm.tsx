import { yupResolver } from '@hookform/resolvers/yup';
import { Button, DatePicker, Heading, Loader, Radio, RadioGroup, TextField, Select as DSSelect, useDatepicker } from '@navikt/ds-react'
import { useRouter } from 'next/router'
import { useContext, useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import * as yup from 'yup'
import { UserState } from '../../../lib/context'
import { useSearchPolly } from '../../../lib/rest/polly';
import Select from 'react-select';
import { SubjectType } from '../../../lib/rest/access';
import { NewAccessRequestDTO, PollyInput } from '../../../lib/rest/generatedDto';
import ErrorStripe from "../../lib/errorStripe";


const tomorrow = () => {
  const date = new Date()
  date.setDate(date.getDate() + 1)
  return date
}

const currentDate = (currentDate: any) => {
  if (typeof currentDate === 'string') return new Date(currentDate)
  else if (currentDate instanceof Date) {
    return currentDate
  }
  return undefined
}

const schema = yup
  .object({
    subject: yup
      .string()
      .required(
        'Du må skrive inn e-postadressen til hvem tilgangen gjelder for'
      )
      .email('E-postadresssen er ikke gyldig'),
    owner: yup
      .string()
      .email('E-postadresssen er ikke gyldig'),
    subjectType: yup
      .mixed<SubjectType>()
      .required('Du må velge hvem tilgangen gjelder for')
      .oneOf([SubjectType.User, SubjectType.Group, SubjectType.ServiceAccount]),
    accessType: yup
      .string()
      .required('Du må velge hvor lenge du ønsker tilgang')
      .oneOf(['eternal', 'until']),
    expires: yup
      .date()
      .nullable()
      .when('accessType', {
        is: 'until',
        then: () => yup.date().required('Du må angi en utløpsdato for tilgang'),
        otherwise: () => yup.date().nullable()
      })
  })
  .required()

export type AccessRequestFormInput = {
  id?: string
  datasetID: string
  expires?: Date
  polly?: PollyInput
  subject?: string
  subjectType?: SubjectType
  owner?: string
  status?: string
  reason?: string
}

interface AccessRequestFormProps {
  accessRequest?: AccessRequestFormInput
  dataset: any
  isEdit: boolean
  onSubmit: (requestData: NewAccessRequestDTO) => Promise<void>
  error: Error | null
  setModal: (value: boolean) => void
}

interface AccessRequestFields {
  subject: string
  subjectType: SubjectType
  owner?: string | undefined
  accessType: string
  expires?: Date | null | undefined
}

const AccessRequestFormV2 = ({
  setModal,
  accessRequest,
  dataset,
  isEdit,
  onSubmit,
  error,
}: AccessRequestFormProps) => {
  const [showSpecifyOwner, setShowSpecifyOwner] = useState(isEdit && accessRequest?.subjectType === SubjectType.ServiceAccount)
  const [searchText, setSearchText] = useState('')
  const [polly, setPolly] = useState<PollyInput | undefined | null>(accessRequest?.polly)
  const [submitted, setSubmitted] = useState(false)
  const router = useRouter()
  const userInfo = useContext(UserState)
  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
    setValue,
    getValues,
  } = useForm({
    resolver: yupResolver(schema),
    defaultValues: {
      subject: accessRequest?.subject ? accessRequest.subject : userInfo?.email ? userInfo.email : "",
      subjectType: accessRequest?.subjectType
        ? accessRequest.subjectType
        : SubjectType.User,
      accessType: !isEdit || accessRequest?.expires ? 'until' : 'eternal',
      expires: isEdit && accessRequest?.expires ? accessRequest.expires : tomorrow(),
    },
  })

  const { datepickerProps, inputProps, selectedDay } = useDatepicker({
    defaultSelected: currentDate(getValues("expires")),
    fromDate: new Date(tomorrow()),
    onDateChange: (d: Date | undefined) => setValue("expires", d),
  });

  const gcpProjects = userInfo?.gcpProjects as any[] || []
  const accessRequestOwner = accessRequest?.owner

  const {
    data: searchResult,
    error: searchError,
    isLoading: loading,
  } = useSearchPolly(searchText)

  const setPollyIfMatches = (input: string) => {
    setPolly(input && searchResult ? searchResult.find((e) => e.externalID === input) : null)
  }
  
  const onSubmitForm = async (data: AccessRequestFields) => {
    setSubmitted(true)
    const accessRequest: NewAccessRequestDTO = {
      datasetID: dataset.id,
      subject: data.subject,
      subjectType: data.subjectType,
      owner: data.owner,
      polly: polly ?? undefined,
      expires: data.accessType === 'until' ? data.expires ? new Date(data.expires).toISOString() : undefined : undefined,
    }
    try {
      await onSubmit(accessRequest)
    }finally{
      setSubmitted(false)
    }
  }

  const getOptionsForSelect = () => {
    const optionsBySearch = searchResult ? searchResult.map((el) => {
      return { value: el.externalID, label: el.name }
    }) : []
    return optionsBySearch
  }

  return (
    <div className="h-full">
      <Heading level="1" size="large" className="pb-8">
        Tilgangssøknad for {dataset.name}
      </Heading>
      <form
        onSubmit={handleSubmit(onSubmitForm)}
        className="flex flex-col gap-10 h-[90%]"
      >
        <div>
          <Controller
            name="subjectType"
            control={control}
            render={({ field }) => (
              <RadioGroup
                {...field}
                legend="Hvem gjelder tilgangen for?"
                error={errors?.subjectType?.message}
                onChange={subjectType => {
                    field.onChange(subjectType);
                    if (subjectType === SubjectType.ServiceAccount) {
                        setShowSpecifyOwner(true)
                        setValue("subject", "")
                    } else if (subjectType === SubjectType.Group) {
                        setShowSpecifyOwner(false)
                        setValue("subject", "")
                    } else {
                        setShowSpecifyOwner(false)
                        setValue("subject", accessRequest?.subject ? accessRequest.subject : userInfo?.email ? userInfo.email : "")
                    }
                }}
              >
                <Radio disabled={isEdit} value={SubjectType.User}>
                  Bruker
                </Radio>
                <Radio disabled={isEdit} value={SubjectType.Group}>
                  Gruppe
                </Radio>
                <Radio disabled={isEdit} value={SubjectType.ServiceAccount}>
                  Servicebruker
                </Radio>
              </RadioGroup>
            )}
          />
          <TextField
            {...register('subject')}
            disabled={isEdit}
            className="hidden-label"
            label="E-post-adresse"
            placeholder="Skriv inn e-post-adresse"
            error={errors?.subject?.message?.toString()}
            size="medium"
          />
          {showSpecifyOwner &&
            <DSSelect
              className="w-full pt-4 gap-2"
              label="Velg eierteam for service bruker"
              {...register('owner', {
              })}
              error={errors.owner?.message?.toString()}
            >
            {accessRequestOwner !== undefined ? <option value={accessRequestOwner} defaultValue={accessRequestOwner}>{accessRequestOwner?.split("@")[0]}</option>
                : <option value="">Velg eiergruppe</option>
            }
            {[
                ...new Set(
                gcpProjects.filter((project) => project.group.email !== accessRequestOwner).map(
                    ({ group }: { group: { name: string } }) => (
                    <option
                        value={
                        userInfo?.googleGroups.filter((g:any) => g.name === group.name)[0]
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
            </DSSelect>
          }
        </div>
        <div>
          <Controller
            name="accessType"
            control={control}
            render={({ field }) => (
              <RadioGroup
                {...field}
                legend="Hvor lenge ønsker du tilgang?"
                error={errors?.accessType?.message}
              >
                <Radio value="until">Til dato</Radio>
                <DatePicker {...datepickerProps}>
                  <DatePicker.Input
                    {...inputProps}
                    label=""
                    disabled={field.value === 'eternal'}
                    error={errors?.expires?.message?.toString()}
                  />
                </DatePicker>
                <Radio value="eternal">For alltid</Radio>
              </RadioGroup>
            )}
          />
          <div>
            <label className="navds-label">
              Velg behandling fra behandlingskatalogen
            </label>
            <Select
              className="basic-single"
              isLoading={loading}
              isClearable={false}
              isRtl={false}
              isSearchable={true}
              name="color"
              defaultValue={polly ? { value: polly.externalID, label: polly.name } : undefined}
              options={getOptionsForSelect()}
              onInputChange={(t: string) => {
                setSearchText(t)
              }}
              onChange={(e: any) => {
                setPollyIfMatches(e.value)
              }}
            />

          </div>
        </div>
        {error && <ErrorStripe error={error} />}
        {submitted && !error && <div>Vennligst vent...<Loader size="small" /></div>}
        <div className="flex flex-row gap-4 grow items-end pb-8">
          <Button
            type="button"
            variant="secondary"
            onClick={() => {
              setModal(false)
              router.push(`/user/requests`)
            }}
          >
            Avbryt
          </Button>
          <Button type="submit" disabled={submitted}>Lagre</Button>
        </div>
      </form>
    </div>
  )
}

export default AccessRequestFormV2
