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
import { UserState } from "../../lib/context"
import DescriptionEditor from '../lib/DescriptionEditor'
import ErrorStripe from "../lib/errorStripe"
import TagsSelector from '../lib/tagsSelector'
import TeamkatalogenSelector from '../lib/teamkatalogenSelector'
import { createMetabaseDashboard } from '../../lib/rest/metabaseDashboards'
import { PublicMetabaseDashboardInput } from '../../lib/rest/generatedDto'
import { HttpError } from '../../lib/rest/request'
import { instanceOf } from 'prop-types'

const schema = yup.object().shape({
    description: yup.string(),
    teamkatalogenURL: yup.string().required('Du må velge team i teamkatalogen'),
    keywords: yup.array(),
    link: yup
        .string()
        .required('Du må legge til en lenke til dashboardet')
        .url('Lenken må være en gyldig URL, fks. https://valid.url.to.page'),
    group: yup.string().required('Du må skrive inn en gruppe for dashboardet')
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
    const [backendError, setBackendError] = useState<Error | undefined>(undefined)
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

    const onSubmit = async (data: any) => {
        const inputData: PublicMetabaseDashboardInput = {
                    description: data.description,
                    keywords: data.keywords,
                    teamkatalogenURL: data.teamkatalogenURL,
                    productAreaID: productAreaID || undefined,
                    teamID: teamID || undefined,
                    link: data.link,
                    group: data.group,
        }

        try {
            setLoading(true)
            await createMetabaseDashboard(inputData)
            setBackendError(undefined)
            router.push('/user/publicDashboards')
        } catch (e) {
            setBackendError(e as Error)
            console.log(e)
        } finally {
          setLoading(false)
        }

    }

    const onCancel = () => {
        router.back()
    }

    const gcpProjects = userData?.gcpProjects as any[] || []

    return (
        <div className="mt-8 md:w-[46rem]">
            <Heading level="1" size="large">
                Legg til public Metabase dashboard
            </Heading>
            <form
                className="pt-12 flex flex-col gap-10"
                onSubmit={handleSubmit(onSubmit)}
            >
                <DescriptionEditor
                    label="Beskrivelse av dashboard"
                    name="description"
                    control={control}
                />
                <TextField
                    className="w-full"
                    label="Lenke til dashboard"
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
                        Lagre
                    </Button>
                </div>
            </form>
        </div>
    )
}
