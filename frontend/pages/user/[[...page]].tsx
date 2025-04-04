import { Checkbox, Tabs } from '@navikt/ds-react'
import Head from 'next/head'
import { useRouter } from 'next/router'
import { useState } from "react"
import { JoinableViewsList } from '../../components/dataProc/joinableViewsList'
import { AccessesList } from '../../components/dataproducts/accessesList'
import ErrorStripe from "../../components/lib/errorStripe"
import InnerContainer from '../../components/lib/innerContainer'
import LoaderSpinner from '../../components/lib/spinner'
import ResultList from '../../components/search/resultList'
import AccessRequestsListForUser from '../../components/user/accessRequests'
import { AccessRequestsForGroup } from '../../components/user/accessRequestsForGroup'
import NadaTokensForUser from '../../components/user/nadaTokens'
import { Workstation } from '../../components/workstation/Workstation'
import { useFetchTokens, useFetchUserData } from '../../lib/rest/userData'

export const UserPages = () => {
    const router = useRouter()
    const [showAllUsersAccesses, setShowAllUsersAccesses] = useState(false)
    const {data, error, isLoading: loading} = useFetchUserData()
    const tokens = useFetchTokens()

    if (error) return <ErrorStripe error={error}/>
    if (loading || !data) return <LoaderSpinner/>

    if (!data)
        return (
            <div>
                <h1>Du må være logget inn!</h1>
                <p>Bruk login-knappen øverst.</p>
            </div>
        )

    const menuItems: Array<{
        title: string
        slug: string
        component: any
    }> = [
        {
            title: 'Mine dataprodukter',
            slug: 'products',
            component: (
                <div className="grid gap-4">
                    <Head>
                        <title>Mine produkter</title>
                    </Head>
                    <h2>Mine produkter</h2>
                    <ResultList dataproducts={data.dataproducts}/>
                </div>
            ),
        },
        {
            title: 'Mine fortellinger',
            slug: 'stories',
            component: (
                <div className="grid gap-4">
                    <Head>
                        <title>Mine fortellinger</title>
                    </Head>
                    <h2>Mine fortellinger</h2>
                    <ResultList stories={data.stories.filter(it => !!it)}/>
                </div>
            ),
        },
        {
            title: 'Mine innsiktsprodukter',
            slug: 'insightProducts',
            component: (
                <div className="grid gap-4">
                    <Head>
                        <title>Mine innsiktsprodukter</title>
                    </Head>
                    <h2>Mine innsiktsprodukter</h2>
                    <ResultList insightProducts={data.insightProducts}/>
                </div>
            ),
        },
        {
            title: 'Tilgangssøknader til meg',
            slug: 'requestsForGroup',
            component: (
                <div className="grid gap-4">
                    <Head>
                        <title>Tilgangssøknader til meg</title>
                    </Head>
                    <h2>Tilgangssøknader til meg</h2>
                    <AccessRequestsForGroup
                        accessRequests={data.accessRequestsAsGranter as any[]}
                    />
                </div>
            ),
        },
        {
            title: 'Mine tilgangssøknader',
            slug: 'requests',
            component: (
                <div className="grid gap-4">
                    <Head>
                        <title>Mine tilgangssøknader</title>
                    </Head>
                    <h2>Mine tilgangssøknader</h2>
                    <AccessRequestsListForUser
                        accessRequests={data.accessRequests}
                    />
                </div>
            ),
        },
        {
            title: 'Mine tilganger',
            slug: 'access',
            component: (
                <div className="grid gap-4">
                    <Head>
                        <title>Mine tilganger</title>
                    </Head>
                    <h2>Mine tilganger</h2>
                    <Tabs
                        defaultValue={router.query.accessCurrentTab ? router.query.accessCurrentTab as string : "owner"}>
                        <Tabs.List>
                            <Tabs.Tab
                                value="owner"
                                label="Eier"
                            />
                            <Tabs.Tab
                                value="granted"
                                label="Innvilgede tilganger"
                            />
                            <Tabs.Tab
                                value="serviceAccountGranted"
                                label="Tilganger servicebrukere"
                            />
                            <Tabs.Tab
                                value="joinable"
                                label="Views tilrettelagt for kobling"
                            />
                        </Tabs.List>
                        <Tabs.Panel value="owner" className="w-full space-y-2 p-4">
                            <AccessesList datasetAccesses={data.accessable.owned}/>
                        </Tabs.Panel>
                        <Tabs.Panel value="granted" className="w-full space-y-2 p-4">
                            <Checkbox onClick={() => setShowAllUsersAccesses(!showAllUsersAccesses)}>Inkluder datasett
                                alle i Nav har tilgang til</Checkbox>
                            <AccessesList datasetAccesses={data.accessable.granted}
                                          showAllUsersAccesses={showAllUsersAccesses}
                                          isRevokable={true}
                            />
                        </Tabs.Panel>
                        <Tabs.Panel value="serviceAccountGranted" className="w-full space-y-2 p-4">
                            <AccessesList datasetAccesses={data.accessable.serviceAccountGranted}
                                          isServiceAccounts={true} 
                                          isRevokable={true}
                            />
                        </Tabs.Panel>
                        <Tabs.Panel value="joinable" className="w-full p-4">
                            <JoinableViewsList/>
                        </Tabs.Panel>
                    </Tabs>
                </div>
            ),
        },
        {
            title: 'Mine team tokens',
            slug: 'tokens',
            component: (
                <div className="grid gap-4">
                    <Head>
                        <title>Mine team tokens</title>
                    </Head>
                    <h2>Mine team tokens</h2>
                    <NadaTokensForUser
                        nadaTokens={tokens.data || []}
                    />
                </div>
            ),
        },
    ]

    if (data.isKnastUser) {
        menuItems.push({
            title: 'Min Knast',
            slug: 'workstation',
            component: (
                <div>
                    <Head>
                        <title>Min Knast</title>
                    </Head>
                    <h2>Min Knast</h2>
                    <Workstation/>
                </div>
            )
        })
    }

    const currentPage = menuItems
        .map((e) => e.slug)
        .indexOf(router.query.page?.[0] ?? 'profile')

    return (
        <InnerContainer>
            <div className="flex flex-row h-full grow pt-8">
                <Head>
                    <title>Brukerside</title>
                </Head>
                <div className="flex flex-col items-stretch justify-between pt-8 w-[20rem]">
                    <div className="flex w-full flex-col gap-2">
                        {menuItems.map(({title, slug}, idx) =>
                            currentPage == idx ? (
                                <p
                                    className="border-l-[6px] border-l-link px-1 font-semibold py-1"
                                    key={idx}
                                >
                                    {title}
                                </p>
                            ) : (
                                <a
                                    className="border-l-[6px] border-l-transparent font-semibold no-underline mx-1 hover:underline hover:cursor-pointer py-1"
                                    href={`/user/${slug}`}
                                    key={idx}
                                >
                                    {title}
                                </a>
                            )
                        )}
                    </div>
                </div>
                {menuItems[currentPage] &&
                    <div className="w-full">{menuItems[currentPage].component}</div>
                }
            </div>
        </InnerContainer>
    )
}

export default UserPages
