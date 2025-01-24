import { BarChartIcon, SidebarLeftIcon } from '@navikt/aksel-icons'
import { Select } from '@navikt/ds-react'
import * as React from 'react'
import { useState } from 'react'
import { PAItems } from '../../pages/productArea/[id]'
import DataproductLogo from '../lib/icons/dataproductLogo'

interface ProductAreaSidebarProps {
    productAreaItems: PAItems
    setCurrentItem: (newCurrent: number) => void
    currentItem: number
    productAreas: any
    selectProductArea: (productAreaId: string) => void
}

const ProductAreaSidebar = ({
    productAreaItems,
    setCurrentItem,
    currentItem,
    productAreas,
    selectProductArea,
}: ProductAreaSidebarProps) => {
    const relevantProductAreas = productAreas
        .filter(
            (it: any) =>
                it.dataproductsNumber ||
                it.storiesNumber ||
                it.insightProductsNumber
        ).sort((l: any, r: any) => (l.name < r.name ? -1 : 1))

    const [collapsed, setCollapsed] = useState(false)
    return (
        <div className={`pr-[2rem] ${collapsed ? 'w-0' : 'w-96'}`}>
            {/*<div className="pr-[2rem] w-96">*/}
            <button className="hidden md:block h-10" onClick={() => setCollapsed(!collapsed)}>
                <SidebarLeftIcon fontSize="1.5rem" title="Vis eller ikke vis sidemeny"></SidebarLeftIcon>
            </button>
            {collapsed ? null : (
                <div className="hidden md:block">
                    <Select
                        className="w-full mb-[1rem]"
                        label=""
                        onChange={(e) => selectProductArea(e.target.value)}
                        value={
                            relevantProductAreas.find((it: any) => it.name == productAreaItems[0].name)
                                ?.id
                        }
                    >
                        {relevantProductAreas.map((it: any, index: number) => (
                            <option key={index} value={it.id}>
                                {it.name}
                            </option>
                        ))}
                    </Select>
                    <div className="flex text-base w-full flex-col gap-2">
                        {productAreaItems.map((d: any, idx: number) =>
                            d.stories.length + d.dataproducts.length + d.insightProducts.length ? (
                                <div
                                    key={idx}
                                    className={`border-l-[6px] py-1 px-2 hover:cursor-default ${currentItem == idx
                                        ? 'border-l-text-action'
                                        : 'border-l-transparent'
                                        }`}
                                >
                                    <a
                                        className="font-semibold no-underline hover:underline"
                                        href="#"
                                        onClick={() => setCurrentItem(idx)}
                                    >
                                        {d.name}
                                    </a>
                                    <div className="flex justify-between w-24">
                                        <span className="flex gap-2 items-center">
                                            <BarChartIcon title="a11y-title" fontSize="1.5rem" />
                                            {' '}
                                            {d.stories.length}
                                        </span>
                                        <span className="flex gap-2 items-center">
                                            <div className="h-[14px] w-[14px] text-text-subtle">
                                                <DataproductLogo />
                                            </div>
                                            {' '}
                                            {d.dataproducts.length}
                                        </span>
                                        <span className="flex gap-2 items-center">
                                            <div className="h-[14px] w-[14px] text-text-subtle">
                                                <BarChartIcon title="a11y-title" fontSize="1.5rem" />

                                            </div>
                                            {' '}
                                            {d.insightProducts.length}
                                        </span>
                                    </div>
                                </div>
                            ) : (
                                <div
                                    key={idx}
                                    className={`border-l-[6px] py-1 px-2 hover:cursor-default ${currentItem == idx
                                        ? 'border-l-text-action'
                                        : 'border-l-transparent'
                                        }`}
                                >
                                    <p className="font-semibold">{d.name}</p>
                                    <div className="flex justify-between w-24">
                                        <span className="flex gap-2 items-center">
                                            <BarChartIcon title="a11y-title" fontSize="1.5rem" />
                                            {' '}
                                            {d.stories.length}
                                        </span>
                                        <span className="flex gap-2 items-center">
                                            <div className="h-[18px] w-[18px] text-text-subtle">
                                                <DataproductLogo />
                                            </div>
                                            {' '}
                                            {d.dataproducts.length}
                                        </span>
                                    </div>
                                </div>
                            )
                        )}
                    </div>
                </div>
            )}
        </div>)
}
export default ProductAreaSidebar
