import { useEffect, useState } from "react"
import { DataproductWithDataset, Dataset, DatasetMap, NewDataproduct, NewDataset, PseudoDataset, UpdateDataproductDto, UpdateDatasetDto } from "./generatedDto"
import { buildCreateDataproductUrl, buildCreateDatasetUrl, buildDeleteDataproductUrl, buildDeleteDatasetUrl, buildGetAccessiblePseudoDatasetsUrl, buildFetchDataproductUrl, buildFetchDatasetUrl, buildMapDatasetToServicesUrl, buildUpdateDataproductUrl, buildUpdateDatasetUrl } from "./apiUrl"
import { deleteTemplate, fetchTemplate, postTemplate, putTemplate } from "./request"

const getDataproduct = async (id: string) =>
    fetchTemplate(buildFetchDataproductUrl(id))

const getDataset = async (id: string) =>
    fetchTemplate(buildFetchDatasetUrl(id))

export const createDataproduct = async (dp: NewDataproduct) =>
    postTemplate(buildCreateDataproductUrl(), dp)

export const updateDataproduct = async (id: string, dp: UpdateDataproductDto) =>
    putTemplate(buildUpdateDataproductUrl(id), dp)

export const deleteDataproduct = async (id: string) =>
    deleteTemplate(buildDeleteDataproductUrl(id))


export const mapDatasetToServices = async (id: string, services: DatasetMap) =>
    postTemplate(buildMapDatasetToServicesUrl(id), services)

export const createDataset = async (dataset: NewDataset) =>
    postTemplate(buildCreateDatasetUrl(), dataset)

export const deleteDataset = async (id: string) =>
    deleteTemplate(buildDeleteDatasetUrl(id))

export const updateDataset = async (id: string, dataset: UpdateDatasetDto) =>
    putTemplate(buildUpdateDatasetUrl(id), dataset)

const getAccessiblePseudoDatasets = async () =>
    fetchTemplate(buildGetAccessiblePseudoDatasetsUrl())

export const useGetDataproduct = (id: string, activeDataSetID?: string) => {
    const [dataproduct, setDataproduct] = useState<DataproductWithDataset | null>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)


    useEffect(() => {
        if (!id) return
        getDataproduct(id).then((res) => res.json())
            .then((dataproduct: DataproductWithDataset) => {
                setError(null)
                setDataproduct(dataproduct)
            })
            .catch((err) => {
                setError(err)
                setDataproduct(null)
            }).finally(() => {
                setLoading(false)
            })
    }, [id, activeDataSetID])

    return { dataproduct, loading, error }
}

export const useGetDataset = (id: string) => {
    const [dataset, setDataset] = useState<Dataset | null>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)

    useEffect(() => {
        if (!id) return
        getDataset(id).then((res) => res.json())
            .then((dataset) => {
                setError(null)
                setDataset(dataset)
            })
            .catch((err) => {
                setError(err)
                setDataset(null)
            }).finally(() => {
                setLoading(false)
            })
    }, [id])

    return { dataset, loading, error }
}

export const useGetAccessiblePseudoDatasets = () => {
    const [accessiblePseudoDatasets, setAccessiblePseudoDatasets] = useState<PseudoDataset[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)


    useEffect(() => {
        getAccessiblePseudoDatasets().then((res) => res.json())
            .then((accessibleds) => {
                setError(null)
                setAccessiblePseudoDatasets(accessibleds)
            })
            .catch((err) => {
                setError(err)
                setAccessiblePseudoDatasets([])
            }).finally(() => {
                setLoading(false)
            })
    }, [])

    return { accessiblePseudoDatasets: accessiblePseudoDatasets, loading, error }
}
