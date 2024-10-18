import { BigqueryColumn, BigQueryTable } from "./generatedDto";
import { fetchTemplate, HttpError } from "./request";
import { buildPath } from "./apiUrl";
import { useQuery } from "react-query";


const bigqueryPath = buildPath('bigquery')
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
    useQuery<string[], HttpError>(['bqDatasets', projectID], ()=>fetchBQDatasets(projectID).then((data)=>data.bqDatasets))

export const useFetchBQTables = (projectID: string, datasetID: string) =>
    useQuery<BigQueryTable[], HttpError>(['bqTables', projectID, datasetID], ()=>fetchBQTables(projectID, datasetID).then((data)=>data.bqTables))

export const useFetchBQcolumns = (projectID: string, datasetID: string, tableID: string) => 
    useQuery<BigqueryColumn[], HttpError>(['bqColumns', projectID, datasetID, tableID], ()=>fetchBQColumns(projectID, datasetID, tableID).then((data)=>data.bqColumns))
