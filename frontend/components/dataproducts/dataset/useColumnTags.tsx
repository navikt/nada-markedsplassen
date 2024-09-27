import { useState } from 'react'
import { useFetchBQcolumns } from '../../../lib/rest/bigquery'

/** BigQueryType defines supported table types in a BigQuery set. */
export enum BigQueryType {
  /** materialized_view is when the table is a BigQuery materialized view. */
  MaterializedView = 'materialized_view',
  /** table is when the table is a regular BigQuery table. */
  Table = 'table',
  /** view is when the table is a BigQuery view. */
  View = 'view'
}

export type PIITagType =
  | 'PII_DirekteIdentifiserende'
  | 'PII_SærligKategori'
  | 'PII_KanVæreIndirekteIdentifiserende'

export const PIITagNames = new Map([
  ['PII_DirekteIdentifiserende', 'Direkte identifiserende'],
  ['PII_SærligKategori', 'Særlig kategori'],
  ['PII_KanVæreIndirekteIdentifiserende', 'Kan være indirekte identifiserende'],
])

export const PIITagOptions = new Map([
  ['PII_DirekteIdentifiserende', 'Direkte identifiserende'],
  ['PII_SærligKategori', 'Særlig kategori'],
  ['PII_KanVæreIndirekteIdentifiserende', 'Ikke klassifisert'],
])

export const DEFAULT_COLUMN_TAG = 'PII_KanVæreIndirekteIdentifiserende'

type TagsMapType = Map<string, Map<string, PIITagType> | undefined>
type PseudoColumnsMapType = Map<string, Map<string, boolean>>

const buildTableKey = (projectID: string, datasetID: string, tableID: string) =>
  `${projectID}.${datasetID}.${tableID}`

export type AnnotateColumnListener = (column: string, tag: PIITagType) => void
export type PseudoColumnListener = (column: string, on: boolean) => void

export const useColumnTags = (
  projectID: string,
  datasetID: string,
  tableID: string,
  dataset?: any
) => {
  const [tagsMap, setTagsMap] = useState<TagsMapType>(
    new Map<string, Map<string, PIITagType>>()
  )

  const [pseudoColumnsMap, setPseudoColumnsMap] = useState<PseudoColumnsMapType>(
    new Map<string, Map<string, boolean>>()
  )

  const fetchColumns = useFetchBQcolumns(projectID, datasetID, tableID)

  var tableKey = buildTableKey(projectID, datasetID, tableID)

  const initDefaultStatesForTable = () => {
    if (
      projectID &&
      datasetID &&
      tableID &&
      !tagsMap.has(tableKey) &&
      fetchColumns.bqColumns &&
      !fetchColumns.error &&
      !fetchColumns.loading &&
      fetchColumns.bqColumns
    ) {
      var newTagsMap = new Map<string, Map<string, PIITagType> | undefined>(
        tagsMap
      )
      var tags = new Map<string, PIITagType>()
      var tagsFromQuery = JSON.parse(dataset?.datasource.piiTags || '{}')

      var newPseudoColumnsMap = new Map<string, Map<string, boolean>>(pseudoColumnsMap)
      var pseudoColumns = new Map<string, boolean>()
      fetchColumns.bqColumns.forEach((it) =>{
        tags.set(
          it.name,
          (!!tagsFromQuery[it.name] &&
            PIITagNames.has(tagsFromQuery[it.name]) &&
            tagsFromQuery[it.name]) ||
            DEFAULT_COLUMN_TAG
        )
        pseudoColumns.set(
          it.name, dataset?.datasource.pseudoColumns?.some((pseudoCol: any)=> it.name === pseudoCol)??false
        )
      }
      )
      newTagsMap.set(tableKey, tags)
      newPseudoColumnsMap.set(tableKey, pseudoColumns)
      setTagsMap(newTagsMap)
      setPseudoColumnsMap(newPseudoColumnsMap)
    }
  }

  initDefaultStatesForTable()

  const annotateColumn = (column: string, tag: PIITagType) => {
    //we cannot directly update the tagsMap and set it back, because the reference to the object
    //will not change, and react do not know it should refresh the components
    var newTagsMap = new Map<string, Map<string, PIITagType> | undefined>(
      tagsMap.set(tableKey, tagsMap.get(tableKey)?.set(column, tag))
    )
    setTagsMap(newTagsMap)
  }

  const selectPseudoColumn = (column: string, on: boolean) => {
    const hasOtherPseudoCol = on || Array.from(pseudoColumnsMap.get(tableKey)??[]).filter(it=> it[0]!== column && it[1]).length
    if(!hasOtherPseudoCol){
      return
    }

    var newPseudoColumnsMap = new Map<string, Map<string, boolean>>(
      pseudoColumnsMap.set(tableKey, pseudoColumnsMap.get(tableKey)?.set(column, on) || new Map<string, boolean>())
    )
    setPseudoColumnsMap(newPseudoColumnsMap)
  }

  return {
    columns:
      !fetchColumns.error && !fetchColumns.loading
        ? fetchColumns.bqColumns
        : undefined,
    loading: fetchColumns.loading,
    error: fetchColumns.error,
    tags: tagsMap.get(tableKey),
    pseudoColumns: pseudoColumnsMap.get(tableKey) || new Map<string, boolean>(),
    annotateColumn: annotateColumn,
    selectPseudoColumn,
  }
}
