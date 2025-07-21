-- name: Search :many
SELECT
	element_id::uuid,
	element_type::text,
	ts_rank_cd(tsv_document, query) AS rank,
	ts_headline('norwegian', "description", query, 'MinWords=10, MaxWords=20, MaxFragments=2 FragmentDelimiter=" … " StartSel="((START))" StopSel="((STOP))"')::text AS excerpt
FROM
	search,
	websearch_to_tsquery('norwegian', @query) query
WHERE
	(
		CASE
			WHEN array_length(@types::text[], 1) > 0 THEN "element_type" = ANY(@types)
			ELSE TRUE
		END
	)
	AND (
		CASE
			WHEN array_length(@keyword::text[], 1) > 0 THEN "keywords" && @keyword
			ELSE TRUE
		END
	)
	AND (
		CASE
			WHEN @query :: text != '' THEN "tsv_document" @@ query
			ELSE TRUE
		END
	)
	AND (
		CASE
			WHEN array_length(@grp::text[], 1) > 0 THEN "group" = ANY(@grp)
			ELSE TRUE
		END
	)
	AND (
		CASE
			WHEN array_length(@team_id::uuid[], 1) > 0 THEN "team_id" = ANY(@team_id)
			ELSE TRUE
		END
	)
ORDER BY rank DESC, created ASC
LIMIT @lim OFFSET @offs;
;

-- name: SearchDatasets :many
SELECT
	ds.*
FROM
	datasets AS ds
	LEFT JOIN datasource_bigquery AS bq ON bq.dataset_id = ds.id
WHERE
 	@keyword ILIKE ANY(ARRAY[ds.name, bq.project_id, bq.dataset, bq.table_name])
	OR concat_ws(
		'.',
		bq.project_id,
		bq.dataset,
		bq.table_name
	) ILIKE '%' || @keyword || '%';