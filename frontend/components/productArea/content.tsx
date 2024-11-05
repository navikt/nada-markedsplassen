import { Tabs } from "@navikt/ds-react";
import { PAItem } from "../../pages/productArea/[id]";
import SearchResultLink from "../search/searchResultLink";
import { useContext } from "react";
import { UserState } from "../../lib/context";

interface ProductAreaContentProps {
    currentItem: PAItem
    currentTab: string
    setCurrentTab: React.Dispatch<React.SetStateAction<string>>
}

const ProductAreaContent = ({ currentItem, currentTab, setCurrentTab }: ProductAreaContentProps) => {
    const userInfo= useContext(UserState)

    return (
        <Tabs
            value={currentTab}
            onChange={setCurrentTab}
            size="medium"
            className="w-full"
        >
            <Tabs.List>
                {currentItem.dashboardURL && <Tabs.Tab
                    value="dashboard"
                    label={currentItem.name}
                />}
                <Tabs.Tab
                    value="stories"
                    label={`Fortellinger (${currentItem.stories.length})`}
                />
                <Tabs.Tab
                    value="products"
                    label={`Produkter (${currentItem.dataproducts.length})`}
                />
                <Tabs.Tab
                    value="insightProducts"
                    label={`Innsiktsprodukt (${currentItem.insightProducts.length})`}
                />                
            </Tabs.List>
            {currentItem.dashboardURL && <Tabs.Panel
                value="dashboard"
                className="h-full w-full py-4"
            >
                <iframe
                    height="4200px"
                    width="100%"
                    src={currentItem.dashboardURL}
                />
            </Tabs.Panel>}
            <Tabs.Panel
                value="stories"
                className="h-full w-full py-4"
            >
                <div className="flex flex-col gap-2">
                    {currentItem.stories && currentItem.stories.map((s: any, idx: number) => (
                        <SearchResultLink
                            resourceType="datafortelling"
                            key={idx}
                            group={s.__typename === 'Story' ? s.owner : {
                                group: s.group,
                                teamkatalogenURL: s.teamkatalogenURL
                            }}
                            name={s.name}
                            description={s.description ? s.description : undefined}
                            keywords={s.keywords}
                            link={`/story/${s.id}`}
                            type={s.__typename}
                            teamkatalogenTeam={s.teamName}
                            productArea={s.productAreaName}
                        />
                    ))}
                    {currentItem.stories.length == 0 && "Ingen fortellinger"}
                </div>
            </Tabs.Panel>
            <Tabs.Panel
                value="products"
                className="h-full w-full py-4"
            >
                <div className="flex flex-col gap-2">
                    {currentItem.dataproducts && currentItem.dataproducts.map((d: any, idx: number) => (
                        <SearchResultLink
                            key={idx}
                            group={d.owner}
                            name={d.name}
                            keywords={d.keywords}
                            description={d.description}
                            link={`/dataproduct/${d.id}/${d.slug}`}
                            teamkatalogenTeam={d.teamName}
                            productArea={d.productAreaName}
                            />
                    ))}
                    {currentItem.dataproducts.length == 0 && "Ingen dataprodukter"}
                </div>
            </Tabs.Panel>
            <Tabs.Panel
                value="insightProducts"
                className="h-full w-full py-4"
            >
                <div className="flex flex-col gap-2">
                    {currentItem.insightProducts && currentItem.insightProducts.map((ip: any, idx: number) => (
                        <SearchResultLink
                            resourceType="innsiktsprodukt"
                            key={idx}
                            group={{
                                teamkatalogenURL: ip.teamkatalogenURL,
                                group: ip.group,
                            }}
                            name={ip.name}
                            keywords={ip.keywords}
                            description={ip.description}
                            link={ip.link}
                            id={ip.id}
                            innsiktsproduktType={ip.type}
                            editable={!!userInfo?.googleGroups?.find((it: any)=> it.email == ip.group)}
                            teamkatalogenTeam={ip.teamName}
                            productArea={ip.productAreaName}
                        />
                    ))}
                    {currentItem.insightProducts.length == 0 && "Ingen innsiktsprodukter"}
                </div>
            </Tabs.Panel>            
        </Tabs>
    )
}

export default ProductAreaContent;
