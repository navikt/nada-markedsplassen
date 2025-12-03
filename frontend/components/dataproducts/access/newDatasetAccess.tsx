import { Button, DatePicker, Heading, Label, Loader, Radio, RadioGroup, TextField, useDatepicker } from "@navikt/ds-react";
import * as yup from 'yup'
import { Controller, useForm} from "react-hook-form";
import { yupResolver } from "@hookform/resolvers/yup";
import { useState } from "react";
import { useRouter } from "next/router";
import { grantDatasetAccess, grantMetabaseAccess, SubjectType } from "../../../lib/rest/access";
import ErrorStripe from "../../lib/errorStripe";

interface NewDatasetAccessProps {
    dataset: any
    setShowNewAccess: (val: boolean) => void
}

enum AccessChoice {
  USER,
  GROUP,
  SERVICE_ACCOUNT,
  ALL_USERS,
}

function accessChoiceToSubjectType(choice: AccessChoice) {
  switch(choice) {
    case AccessChoice.USER: return SubjectType.User
    case AccessChoice.SERVICE_ACCOUNT: return SubjectType.ServiceAccount
    case AccessChoice.GROUP:
    case AccessChoice.ALL_USERS: return SubjectType.Group
  }
}

const tomorrow = () => {
    const date = new Date()
    date.setDate(date.getDate() + 1)
    return date
}

const schema = yup
  .object({
    accessChoice: yup
      .number()
      .required('Du må velge hvem tilgangen gjelder for')
      .oneOf([AccessChoice.USER, AccessChoice.GROUP, AccessChoice.SERVICE_ACCOUNT, AccessChoice.ALL_USERS]),
    subject: yup
      .string()
      .trim()
      .when('accessChoice', {
        is: (val: AccessChoice) => val !== AccessChoice.ALL_USERS, 
        then: (s) => s.required('Du må skrive inn e-postadressen til hvem tilgangen gjelder for').email('E-postadresssen er ikke gyldig'),
        otherwise: (s) => s,
      }),
    owner: yup
      .string()
      .trim()
      .email(),
    accessType: yup
      .string()
      .required('Du må velge hvor lenge du ønsker tilgang')
      .oneOf(['eternal', 'until']),
    expires: yup
      .string()
      .nullable()
      .when('accessType', {
        is: 'until',
        then: () => yup.string().nullable().matches(/\d{4}-[01]\d-[0-3]\d/, 'Du må velge en dato')
      })
      ,
  })
  .required()

const NewDatasetAccess = ({dataset, setShowNewAccess}: NewDatasetAccessProps) => {
    const [error, setError] = useState<any>(null)
    const [submitted, setSubmitted] = useState(false)
    const [showSpecifyOwner, setShowSpecifyOwner] = useState(false)
    const [showEmail, setShowEmail] = useState(true)
    const router = useRouter()
    const {
        register,
        handleSubmit,
        control,
        formState: { errors },
        setValue
      } = useForm({
        resolver: yupResolver(schema),
        defaultValues: {
          subject: '',
          accessChoice: AccessChoice.USER,
          accessType: 'until',
          expires: '',
        },
      })

    const { datepickerProps, inputProps } = useDatepicker({
      fromDate: tomorrow(),
      onDateChange: (d: Date | undefined) => setValue("expires", d ? d.toISOString() : ''),
    });

    const onSubmitForm = async (requestData: any) => {
        setSubmitted(true)
        requestData.datasetID = dataset.id
        try{
          await grantDatasetAccess({
            datasetID: dataset.id /* uuid */,
            expires: requestData.accessType === "until" ? new Date(requestData.expires).toISOString() : undefined /* RFC3339 */,
            subject: requestData.accessChoice === AccessChoice.ALL_USERS ? "all-users@nav.no" : requestData.subject.trim(),
            owner: (requestData.owner !== "" || undefined) && requestData.subjectType === SubjectType.ServiceAccount ? requestData.owner.trim(): requestData.subject.trim(),
            subjectType: accessChoiceToSubjectType(requestData.accessChoice),
          })

          
          if (dataset.metabaseDataset && dataset.metabaseDataset.Type !== "open" && requestData.accessChoice === AccessChoice.USER) {
            await grantMetabaseAccess({
              datasetID: dataset.id /* uuid */,
              expires: undefined, 
              subject: requestData.subject.trim(),
              owner: (requestData.owner !== "" || undefined) && requestData.subjectType === SubjectType.ServiceAccount ? requestData.owner.trim(): requestData.subject.trim(),
              subjectType: accessChoiceToSubjectType(requestData.accessChoice),
            })
          }
        }catch(e){
            setError(e)
        }

        router.reload()
    }

    return (
        <div className="h-full">
      <Heading level="1" size="large" className="pb-8">
        Legg til tilgang for {dataset.name}
      </Heading>
      <form
        onSubmit={handleSubmit(onSubmitForm)}
        className="flex flex-col gap-10 h-[90%]"
      >
        <div>
          <Controller
            name="accessChoice"
            control={control}
            render={({ field }) => (
              <RadioGroup
                {...field}
                legend="Hvem gjelder tilgangen for?"
                error={errors?.accessChoice?.message}
                onChange={accessChoice => {
                    field.onChange(accessChoice);
                    if (accessChoice === AccessChoice.SERVICE_ACCOUNT) {
                        setShowSpecifyOwner(true)
                    } else {
                        setShowSpecifyOwner(false)
                    }
                    if (accessChoice === AccessChoice.ALL_USERS) {
                        setValue("subject", "all-users@nav.no")
                        setShowEmail(false)
                    } else {
                        setValue("subject", "")
                        setShowEmail(true)
                    }

                }}
              >
                <Radio value={AccessChoice.USER}>
                  Bruker
                </Radio>
                <Radio value={AccessChoice.GROUP}>
                  Gruppe
                </Radio>
                <Radio value={AccessChoice.SERVICE_ACCOUNT}>
                  Servicebruker
                </Radio>
                <Radio value={AccessChoice.ALL_USERS}>
                  Alle i Nav
                </Radio>
              </RadioGroup>
            )}
          />
        {showEmail &&
          <TextField
            {...register('subject')}
            className="hidden-label"
            label="E-post-adresse"
            placeholder="Skriv inn e-post-adresse"
            error={errors?.subject?.message}
            size="medium"
          />
        }
        {showSpecifyOwner && 
            <div className="flex flex-col gap-1 pt-2">
                <Label>Eierteam</Label>
                <TextField
                    {...register('owner')}
                    className="hidden-label"
                    label="E-post-adresse"
                    placeholder="Skriv inn e-post-adresse"
                    error={errors?.subject?.message}
                    size="medium"
                />
            </div>
        }
        </div>
        <div>
          <Controller
            name="accessType"
            control={control}
            render={({ field }) => (
              <RadioGroup
                {...field}
                legend="Hvor lenge skal tilgangen vare?"
                error={errors?.accessType?.message}
              >
                <Radio value="until">Til dato</Radio>
                <DatePicker {...datepickerProps}>
                  <DatePicker.Input 
                    {...inputProps} 
                    label="" 
                    disabled={field.value === 'eternal'} 
                    error={errors?.expires?.message} 
                  />
                </DatePicker>
                <Radio value="eternal">For alltid</Radio>
              </RadioGroup>
            )}
          />
        </div>
        { error && <ErrorStripe error={error} /> }
        {submitted && !error && <div>Vennligst vent...<Loader size="small"/></div>}
        <div className="flex flex-row gap-4 grow items-end">
          <Button
            type="button"
            variant="secondary"
            onClick={() => {setShowNewAccess(false)}}
          >
            Avbryt
          </Button>
          <Button type="submit" disabled={submitted}>Lagre</Button>
        </div>
      </form>
    </div>
    )
}

export default NewDatasetAccess;
