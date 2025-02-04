import { useEffect, useState } from "react";
import { ProductArea, ProductAreasDto, ProductAreaWithAssets } from "./generatedDto";
import { fetchTemplate, HttpError } from "./request";
import { buildUrl } from "./apiUrl";
import { useQueries, useQuery } from '@tanstack/react-query';

const productAreasPath = buildUrl('productareas')
const buildGetProductAreasUrl = () => productAreasPath()()
const buildGetProductAreaUrl = (id: string) => productAreasPath(id)()

const getProductAreas = async () => 
    fetchTemplate(buildGetProductAreasUrl())

const getProductArea = async (id: string) => 
    fetchTemplate(buildGetProductAreaUrl(id))

const enrichProductArea = (productArea: ProductArea) => {
    return {
        ...productArea,
        dataproductsNumber: productArea.teams.reduce((acc: number, t: any) => acc + t.dataproductsNumber, 0),
        storiesNumber: productArea.teams.reduce((acc: number, t: any) => acc + t.storiesNumber, 0),
        insightProductsNumber: productArea.teams.reduce((acc: number, t: any) => acc + t.insightProductsNumber, 0),
    }

}

export const useGetProductAreas = () => useQuery<ProductArea[], HttpError>({
    queryKey: ['productAreas'], 
    queryFn: ()=>
    getProductAreas().then((productAreaDto: ProductAreasDto) => productAreaDto.productAreas.filter(it=> !!it).map(enrichProductArea))})

const enrichProductAreaWithAssets = (productArea: ProductAreaWithAssets) => {
    return {
        ...productArea,
        dataproducts: productArea.teamsWithAssets.flatMap((t: any) => t.dataproducts),
        stories: productArea.teamsWithAssets.flatMap((t: any) => t.stories),
        insightProducts: productArea.teamsWithAssets.flatMap((t: any) => t.insightProducts),
    }

}

export const useGetProductArea = (id: string) => useQuery<ProductAreaWithAssets, HttpError>({
    queryKey: ['productArea', id], 
    queryFn: ()=>
    getProductArea(id).then((productAreaDto: ProductAreaWithAssets) => enrichProductAreaWithAssets(productAreaDto))
})
