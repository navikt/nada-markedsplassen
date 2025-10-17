import { yupResolver } from '@hookform/resolvers/yup'
import {
    Button,
    Checkbox,
    Heading,
    Select,
    TextField,
} from '@navikt/ds-react'
import { useRouter } from 'next/router'
import { useContext, useState } from 'react'
import { useForm } from 'react-hook-form'
import * as yup from 'yup'
import { UserState } from "../../lib/context"
import { createInsightProduct } from '../../lib/rest/insightProducts'
import DescriptionEditor from '../lib/DescriptionEditor'
import ErrorStripe from "../lib/errorStripe"
import TagsSelector from '../lib/tagsSelector'
import TeamkatalogenSelector from '../lib/teamkatalogenSelector'

const schema = yup.object().shape({
    name: yup.string().nullable().required('Skriv inn navnet på innsiktsproduktet'),
    description: yup.string(),
    teamkatalogenURL: yup.string().required('Du må velge team i teamkatalogen'),
    keywords: yup.array(),
    type: yup.string(),
    link: yup
        .string()
        .required('Du må legge til en lenke til innsiktsproduktet')
        .url('Lenken må være en gyldig URL, fks. https://valid.url.to.page'), // Add this line to validate the link as a URL    type: yup.string().required('Du må velge en type for innsiktsproduktet'),
    group: yup.string().required('Du må skrive inn en gruppe for innsiktsproduktet')
})

export const NewInsightProductForm = () => {
    const router = useRouter()
    const [productAreaID, setProductAreaID] = useState<string>('')
    const [teamID, setTeamID] = useState<string>('')
    const userData = useContext(UserState)
    const [isPrivacyCheckboxChecked, setIsPrivacyCheckboxChecked] = useState(false)
    const [backendError, setBackendError] = useState<Error | undefined>(undefined)

    const handlePrivacyCheckboxChange = () => {
        setIsPrivacyCheckboxChecked(!isPrivacyCheckboxChecked)
    }

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
            name: undefined,
            description: '',
            teamkatalogenURL: '',
            keywords: [] as string[],
            link: '',
            type: '',
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
        const inputData = {
                    name: data.name,
                    description: data.description,
                    keywords: data.keywords,
                    teamkatalogenURL: data.teamkatalogenURL,
                    productAreaID: productAreaID || undefined,
                    teamID: teamID || undefined,
                    link: data.link,
                    type: data.type,
                    group: data.group,
        }

        try {
            await createInsightProduct(inputData)
            setBackendError(undefined)
            router.push('/user/insightProducts')
        } catch (e) {
            setBackendError(e as Error)
            console.log(e)
        }

    }

    const onCancel = () => {
        router.back()
    }

    const gcpProjects = userData?.gcpProjects as any[] || []

    return (
        <div className="mt-8 md:w-[46rem]">
            <Heading level="1" size="large">
                Legg til innsiktsprodukt
            </Heading>
            <form
                className="pt-12 flex flex-col gap-10"
                onSubmit={handleSubmit(onSubmit)}
            >
                <TextField
                    className="w-full"
                    label="Navn på innsiktsprodukt"
                    {...register('name')}
                    error={errors.name?.message?.toString()}
                />
                <DescriptionEditor
                    label="Beskrivelse av hva innsiktsproduktet kan brukes til"
                    name="description"
                    control={control}
                />
                <TextField
                    className="w-full"
                    label="Lenke til innsiktsprodukt"
                    {...register('link')}
                    error={errors.link?.message?.toString()}
                />
                <Select
                    className="w-full"
                    label="Velg type innsiktsprodukt"
                    {...register('type')}
                    error={errors.type?.message?.toString()}
                >
                    <option value="">Velg type</option>
                    <option value="Amplitude">Amplitude</option>
                    <option value="Grafana">Grafana</option>
                    <option value="HotJar">HotJar</option>
                    <option value="TaskAnalytics">TaskAnalytics</option>
                    <option value="Metabase">Metabase</option>
                    <option value="Tableau">Tableau</option>
                    <option value="Annet">Annet</option>
                </Select>
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
                <div className="flex items-center mt-4">
                    <Checkbox
                        size="small"
                        checked={isPrivacyCheckboxChecked}
                        onChange={handlePrivacyCheckboxChange}
                        className="pl-2"
                    >
                        Innholdsprodukter inneholder ikke personsensitive eller identifiserende opplysninger
                    </Checkbox>
                </div>
                <div className="flex flex-row gap-4 mb-16">
                    <Button type="button" variant="secondary" onClick={onCancel}>
                        Avbryt
                    </Button>
                    <Button type="submit" disabled={!isPrivacyCheckboxChecked}>
                        Lagre
                    </Button>
                </div>
            </form>
        </div>
    )
}
