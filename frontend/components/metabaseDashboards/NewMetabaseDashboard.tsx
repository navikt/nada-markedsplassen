import { yupResolver } from '@hookform/resolvers/yup'
import {
    BodyShort,
    Button,
    Heading,
    Select,
    TextField,
} from '@navikt/ds-react'
import { useRouter } from 'next/router'
import { useContext, useState } from 'react'
import { useForm } from 'react-hook-form'
import * as yup from 'yup'
import { UserState } from "../../lib/context"
import { PublicMetabaseDocumentInput } from '../../lib/rest/generatedDto'
import { createMetabaseDashboard } from '../../lib/rest/metabaseDashboards'
import { createMetabaseDocument } from '../../lib/rest/metabaseDocuments'
import { HttpError } from '../../lib/rest/request'
import DescriptionEditor from '../lib/DescriptionEditor'
import ErrorStripe from "../lib/errorStripe"
import TagsSelector from '../lib/tagsSelector'
import TeamkatalogenSelector from '../lib/teamkatalogenSelector'

const isDocumentLink = (link: string | undefined) => {
    try {
        return new URL(link || '').pathname.includes('/document/')
    } catch {
        return false
    }
}

// yup's url() rejects localhost, which is needed for local development.
const isValidHttpUrl = (link: string | undefined) => {
    try {
        const parsed = new URL(link || '')
        return parsed.protocol === 'http:' || parsed.protocol === 'https:'
    } catch {
        return false
    }
}

const schema = yup.object().shape({
    description: yup.string(),
    teamkatalogenURL: yup.string().required('Du må velge team i teamkatalogen'),
    keywords: yup.array(),
    link: yup
        .string()
        .trim()
        .required('Du må legge til en lenke')
        .test('http-url', 'Lenken må være en gyldig URL, f.eks. https://valid.url.to.page', (value) => !value || isValidHttpUrl(value)),
    group: yup.string().required('Du må velge en eiergruppe')
})

export type FormValues = {
    description?: string | undefined
    teamkatalogenURL: string
    keywords?: any[] | undefined
    link: string
    group: string
}

export const NewMetabaseDashboardForm = () => {
    const router = useRouter()
    const [productAreaID, setProductAreaID] = useState<string>('')
    const [teamID, setTeamID] = useState<string>('')
    const userData = useContext(UserState)
    const [backendError, setBackendError] = useState<HttpError | undefined>(undefined)
    const [isLoading, setLoading] = useState(false)

    const {
        register,
        handleSubmit,
        watch,
        formState,
        setValue,
        control,
    } = useForm<FormValues>({
        resolver: yupResolver<FormValues, any, any>(schema),
        defaultValues: {
            description: '',
            teamkatalogenURL: '',
            keywords: [] as string[],
            link: '',
            group: '',
        },
    })

    const { errors } = formState
    const keywords = watch('keywords')

    const onDeleteKeyword = (keyword: string) => {
        keywords !== undefined ? 
        setValue('keywords', keywords.filter((k: string) => k !== keyword))
        :
        setValue('keywords', [])
    }

    const onAddKeyword = (keyword: string) => {
        keywords
            ? setValue('keywords', [...keywords, keyword])
            : setValue('keywords', [keyword])
    }

    const onSubmit = async (data: FormValues) => {
        setLoading(true)
        const inputData: PublicMetabaseDocumentInput = {
            description: data.description,
            keywords: data.keywords || [],
            teamkatalogenURL: data.teamkatalogenURL,
            productAreaID: productAreaID || undefined,
            teamID: teamID || undefined,
            link: data.link,
            group: data.group,
        }

        try {
            if (isDocumentLink(data.link)) {
                await createMetabaseDocument(inputData)
            } else {
                await createMetabaseDashboard(inputData)
            }
            setBackendError(undefined)
            router.push('/user/publicDashboards')
        } catch (e) {
            setBackendError(e as HttpError)
        } finally {
            setLoading(false)
        }

    }

    const onCancel = () => {
        router.back()
    }

    const gcpProjects = userData?.gcpProjects as any[] || []

    return (
        <div className="mt-8 ax-md:w-[46rem]">
            <Heading level="1" size="large">
                Legg til ny offentlig Metabase-lenke
            </Heading>
            <BodyShort className="mt-4">
                Når du oppretter en offentlig Metabase-lenke, blir innholdet tilgjengelig uten innlogging. Det gjelder også resultater fra innebygde spørsmål. Du har ansvar for at innholdet ikke inneholder personopplysninger.
            </BodyShort>
            <form
                className="pt-12 flex flex-col gap-10"
                onSubmit={handleSubmit(onSubmit)}
            >
                <DescriptionEditor
                    label="Beskrivelse"
                    name="description"
                    control={control}
                />
                <TextField
                    className="w-full"
                    label="Lenke til dashboard eller dokument"
                    {...register('link')}
                    error={errors.link?.message?.toString()}
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
                    gcpGroups={userData?.gcpProjects?.map((it: any) => it.group.email)}
                    register={register}
                    watch={watch}
                    errors={errors}
                    setValue={setValue}
                    setProductAreaID={setProductAreaID}
                    setTeamID={setTeamID}
                />
                <TagsSelector
                    onAdd={onAddKeyword}
                    onDelete={onDeleteKeyword}
                    tags={keywords || []}
                />
                {backendError && <ErrorStripe error={backendError} />}
                <div className="flex flex-row gap-4 mb-16">
                    <Button type="button" variant="secondary" onClick={onCancel} disabled={isLoading}>
                        Avbryt
                    </Button>
                    <Button type="submit" loading={isLoading}>
                        Opprett lenke
                    </Button>
                </div>
            </form>
        </div>
    )
}
