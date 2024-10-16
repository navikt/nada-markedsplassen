import { useEffect, useState } from "react";
import { ProductArea, ProductAreasDto, ProductAreaWithAssets } from "./generatedDto";
import { fetchTemplate } from "./request";
import { buildGetProductAreasUrl, buildGetProductAreaUrl } from "./apiUrl";

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

export const useGetProductAreas = () => {
    const [productAreas, setProductAreas] = useState<ProductArea[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    useEffect(() => {
        getProductAreas().then((res) => res.json())
            .then((productAreaDto: ProductAreasDto) => {
            setError(null);
            setProductAreas([...productAreaDto.productAreas.filter(it=> !!it).map(enrichProductArea)]);
        })
            .catch((err) => {
            setError(err);
            setProductAreas([]);
        }).finally(() => {
            setLoading(false);
        });
    }, []);
    return { productAreas, loading, error };
}

const enrichProductAreaWithAssets = (productArea: ProductAreaWithAssets) => {
    return {
        ...productArea,
        dataproducts: productArea.teamsWithAssets.flatMap((t: any) => t.dataproducts),
        stories: productArea.teamsWithAssets.flatMap((t: any) => t.stories),
        insightProducts: productArea.teamsWithAssets.flatMap((t: any) => t.insightProducts),
    }

}

export const useGetProductArea = (id: string) => {
    const [productArea, setProductArea] = useState<ProductAreaWithAssets|null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<any>(undefined);
    useEffect(() => {
        getProductArea(id).then((res) => res.json())
            .then((productAreaDto: ProductAreaWithAssets) => {
            setError(undefined);
            setProductArea(enrichProductAreaWithAssets(productAreaDto));
        })
            .catch((err) => {
            setError({
                message: `Failed to fetch product area, please check the product area ID: ${err.message}`,
                status: err.status
            });
            setProductArea(null);
        }).finally(() => {
            setLoading(false);
        });
    }, [id]);
    return { productArea, loading, error };
}