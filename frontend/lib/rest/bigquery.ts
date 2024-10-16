import { useEffect, useState } from "react"
import { BigqueryColumn, BigQueryTable, BQDatasets, BQTables } from "./generatedDto"
import { fetchTemplate } from "./request"
import { buildFetchBQColumnsUrl, buildFetchBQDatasetsUrl, buildFetchBQTablesUrl } from "./apiUrl"

const fetchBQDatasets = async (projectID: string) => 
    fetchTemplate(buildFetchBQDatasetsUrl(projectID))

const fetchBQTables = async (projectID: string, datasetID: string) => 
    fetchTemplate(buildFetchBQTablesUrl(projectID, datasetID))
    
const fetchBQColumns = async (projectID: string, datasetID: string, tableID: string) => 
    fetchTemplate(buildFetchBQColumnsUrl(projectID, datasetID, tableID))


//TODO: empty project id, dataset id, table id

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