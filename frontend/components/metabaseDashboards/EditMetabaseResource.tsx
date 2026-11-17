import { yupResolver } from '@hookform/resolvers/yup'
import {
    Button,
    Heading,
    TextField,
} from '@navikt/ds-react'
import { useRouter } from 'next/router'
import { useContext, useState } from 'react'
import { useForm } from 'react-hook-form'
import * as yup from 'yup'
import { UserState } from "../../lib/context"
import { PublicMetabaseDashboardEditInput } from '../../lib/rest/generatedDto'
import { updateMetabaseDashboard } from '../../lib/rest/metabaseDashboards'
import { updateMetabaseDocument } from '../../lib/rest/metabaseDocuments'
import DescriptionEditor from '../lib/DescriptionEditor'
import ErrorStripe from "../lib/errorStripe"
import TagsSelector from '../lib/tagsSelector'
import TeamkatalogenSelector from '../lib/teamkatalogenSelector'

export type MetabaseResourceType = 'dashboard' | 'document'

const resources: {
    [key in MetabaseResourceType]: {
        label: string
        update: (id: string, input: PublicMetabaseDashboardEditInput) => Promise<unknown>
    }
} = {
    dashboard: { label: 'dashboard', update: updateMetabaseDashboard },
    document: { label: 'dokument', update: updateMetabaseDocument },
}

const schema = yup.object().shape({
    name: yup.string().nullable().required('Skriv inn navnet'),
    description: yup.string(),
    teamkatalogenURL: yup.string().required('Du må velge team i teamkatalogen'),
    keywords: yup.array(),
})

export interface EditMetabaseResourceProps {
    resourceType: MetabaseResourceType
    id: string
    name: string
    description?: string
    keywords: string[]
    teamkatalogenURL: string
    group: string
    link: string
    teamID?: string
}

export type FormValues = {
    name: string | null
    description?: string | undefined
    teamkatalogenURL: string
    keywords?: any[] | undefined
}

export const EditMetabaseResource = ({ resourceType, id, name, description, keywords, teamkatalogenURL, link }: EditMetabaseResourceProps) => {
    const router = useRouter()
    const [productAreaID, setProductAreaID] = useState<string>('')
    const [teamID, setTeamID] = useState<string>('')
    const userInfo = useContext(UserState)
    const [error, setError] = useState<Error | undefined>(undefined)
    const [loading, setLoading] = useState(false)
    const { label, update } = resources[resourceType]

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
            name: name,
            description: description,
            keywords: keywords,
            teamkatalogenURL: teamkatalogenURL,
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
        const editData = {
            description: data.description,
            keywords: data.keywords,
            teamkatalogenURL: data.teamkatalogenURL,
            productAreaID: productAreaID || undefined,
            teamID: teamID || undefined,
        }

        setLoading(true)
        update(id, editData).then(() => {
            setError(undefined)
            router.push('/user/publicDashboards')
        }).catch(e => {
            console.log(e)
            setError(e)
        }).finally(() => {
            setLoading(false)
        })
    }

    return (
        <div className="mt-8 ax-md:w-[46rem]">
            <Heading level="1" size="large">
                Rediger Metabase-{label}
            </Heading>
            <form
                className="pt-12 flex flex-col gap-10"
                onSubmit={handleSubmit(onSubmit)}
            >
                <TextField
                    className="w-full"
                    label={`Navn på Metabase-${label}`}
                    disabled
                    value={name}
                    error={errors.name?.message?.toString()}
                />
                <DescriptionEditor
                    label={`Beskrivelse av hva Metabase-${label}et kan brukes til`}
                    name="description"
                    control={control}
                />
                <TeamkatalogenSelector
                    gcpGroups={userInfo?.gcpProjects.map((it: any) => it.group.email)}
                    register={register}
                    watch={watch}
                    errors={errors}
                    setValue={setValue}
                    setProductAreaID={setProductAreaID}
                    setTeamID={setTeamID}
                />
                <TextField
                    className="w-full"
                    disabled
                    label={`Lenke til Metabase-${label}`}
                    value={link}
                />
                <TagsSelector
                    onAdd={onAddKeyword}
                    onDelete={onDeleteKeyword}
                    tags={kw || []}
                />
                {error && <ErrorStripe error={error} />}
                <div className="flex flex-row gap-4 mb-16">
                    <Button type="button" variant="secondary" onClick={() => router.push('/user/publicDashboards')}>
                        Avbryt
                    </Button>
                    <Button type="submit" disabled={loading}>Lagre</Button>
                </div>
            </form>
        </div>
    )
}
