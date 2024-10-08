import { useEffect, useState } from "react"
import { fetchBQColumnsUrl, fetchBQDatasetsUrl, fetchBQTablesUrl, fetchTemplate } from "./restApi"
import { BigqueryColumn, BigQueryTable, BQDatasets, BQTables } from "./generatedDto"

const fetchBQDatasets = async (projectID: string) => {
    if(!projectID) return Promise.resolve({} as Response)
    const url = fetchBQDatasetsUrl(projectID)
    return fetchTemplate(url)
}

const fetchBQTables = async (projectID: string, datasetID: string) => {
    if(!projectID || !datasetID) return Promise.resolve({} as Response)
    const url = fetchBQTablesUrl(projectID, datasetID)
    return fetchTemplate(url)
}

const fetchBQColumns = async (projectID: string, datasetID: string, tableID: string) => {
    if(!projectID || !datasetID || !tableID) return Promise.resolve({} as Response)
    const url = fetchBQColumnsUrl(projectID, datasetID, tableID)
    return fetchTemplate(url)
}

export const useFetchBQDatasets = (projectID: string) => {
    const [datasets, setDatasets] = useState<string[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<any>(null)

    useEffect(() => {
        setLoading(true)
        fetchBQDatasets(projectID).then((res) => res.json())
            .then((data) => {
                setError(null)
                setDatasets(data.bqDatasets)
            })
            .catch((err) => {
                setError(err)
                setDatasets([])
            }).finally(() => {
                setLoading(false)
            })
    }, [projectID])

    return { bqDatasets: datasets, loading, error }
}

export const useFetchBQTables = (projectID: string, datasetID: string) => {
    const [tables, setTables] = useState<BigQueryTable[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<any>(null)

    useEffect(() => {
        setLoading(true)
        fetchBQTables(projectID, datasetID).then((res) => res.json())
            .then((data) => {
                setError(null)
                setTables(data.bqTables)
            })
            .catch((err) => {
                setError(err)
                setTables([])
            }).finally(() => {
                setLoading(false)
            })
    }, [projectID, datasetID])

    return { bqTables: tables, loading, error }
}

export const useFetchBQcolumns = (projectID: string, datasetID: string, tableID: string) => {
    const [columns, setColumns] = useState<BigqueryColumn[]|undefined>(undefined)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<any>(null)

    useEffect(() => {
        setLoading(true)
        fetchBQColumns(projectID, datasetID, tableID).then((res) => res.json())
            .then((data) => {
                setError(null)
                setColumns(data.bqColumns)
            })
            .catch((err) => {
                setError(err)
                setColumns([])
            }).finally(() => {
                setLoading(false)
            })
    }, [projectID, datasetID, tableID])

    return { bqColumns: columns, loading, error }
}