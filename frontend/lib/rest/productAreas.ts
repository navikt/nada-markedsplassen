import { useEffect, useState } from "react";
import { fetchTemplate, getProductAreaUrl, getProductAreasUrl } from "./restApi";
import { ProductArea, ProductAreasDto, ProductAreaWithAssets } from "./generatedDto";

const getProductAreas = async () => {
    const url = getProductAreasUrl();
    return fetchTemplate(url)
}

const getProductArea = async (id: string) => {
    const url = getProductAreaUrl(id)
    return fetchTemplate(url)
}

const enrichProductArea = (productArea: any) => {
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
            setProductAreas([...productAreaDto.productAreas.map(enrichProductArea)]);
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

const enrichProductAreaWithAssets = (productArea: any) => {
    return {
        ...productArea,
        dataproducts: productArea.teams.flatMap((t: any) => t.dataproducts),
        stories: productArea.teams.flatMap((t: any) => t.stories),
        insightProducts: productArea.teams.flatMap((t: any) => t.insightProducts),
    }

}

export const useGetProductArea = (id: string) => {
    const [productArea, setProductArea] = useState<ProductAreaWithAssets|null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<any>(undefined);
    useEffect(() => {
        getProductArea(id).then((res) => res.json())
            .then((productAreaDto) => {
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