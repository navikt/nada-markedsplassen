import { BigqueryColumn, BigQueryTable } from "./generatedDto";
import { fetchTemplate, HttpError } from "./request";
import { buildUrl } from "./apiUrl";
import { useQuery } from '@tanstack/react-query';


const bigqueryPath = buildUrl('bigquery')
const buildFetchBQDatasetsUrl = (projectId: string) => bigqueryPath('datasets')({projectId: projectId})
const buildFetchBQTablesUrl = (projectId: string, datasetId: string) => bigqueryPath('tables')({projectId: projectId, datasetId: datasetId})
const buildFetchBQColumnsUrl = (projectId: string, datasetId: string, tableId: string) =>  bigqueryPath('columns')({projectId: projectId, datasetId: datasetId, tableId: tableId})

const fetchBQDatasets = async (projectID: string) => 
    fetchTemplate(buildFetchBQDatasetsUrl(projectID))

const fetchBQTables = async (projectID: string, datasetID: string) => 
    fetchTemplate(buildFetchBQTablesUrl(projectID, datasetID))
    
const fetchBQColumns = async (projectID: string, datasetID: string, tableID: string) => 
    fetchTemplate(buildFetchBQColumnsUrl(projectID, datasetID, tableID))


export const useFetchBQDatasets = (projectID: string) => 
    useQuery<string[], HttpError>({
        queryKey: ['bqDatasets', projectID], 
        queryFn: ()=>fetchBQDatasets(projectID).then((data)=>data.bqDatasets)
    })

export const useFetchBQTables = (projectID: string, datasetID: string) =>
    useQuery<BigQueryTable[], HttpError>({
        queryKey: ['bqTables', projectID, datasetID], 
        queryFn: ()=>fetchBQTables(projectID, datasetID).then((data)=>data.bqTables)})

export const useFetchBQcolumns = (projectID: string, datasetID: string, tableID: string) => 
    useQuery<BigqueryColumn[], HttpError>({
        queryKey: ['bqColumns', projectID, datasetID, tableID], 
        queryFn: ()=>fetchBQColumns(projectID, datasetID, tableID).then((data)=>data.bqColumns)
    })
