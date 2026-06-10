import { Alert, BodyShort, Label, ReadMore, Tabs } from '@navikt/ds-react'
import { BigQuery, ViewTable } from '../../lib/rest/generatedDto'
import Copy from '../lib/copy'

interface MetabaseGrantSnippetProps {
  saEmail: string
  datasource: BigQuery
}

interface SnippetBlockProps {
  label: string
  code: string
}

const SnippetBlock = ({ label, code }: SnippetBlockProps) => (
  <div className="flex flex-col gap-1">
    <Label size="small">{label}</Label>
    <div className="relative">
      <pre className="bg-ax-bg-neutral-moderate rounded px-3 py-2 pr-10 text-sm font-mono overflow-x-auto whitespace-pre-wrap break-all">
        {code}
      </pre>
      <div className="absolute top-2 right-2">
        <Copy text={code} />
      </div>
    </div>
  </div>
)

const MetabaseGrantSnippet = ({ saEmail, datasource }: MetabaseGrantSnippetProps) => {
  const { projectID: project, dataset, table, tableType } = datasource
  const isView = tableType === ViewTable

  const dbtSnippet =
    `config:
  grants:
    "roles/bigquery.dataViewer":
      - "serviceAccount:${saEmail}"`

  const terraformSnippet =
    `# Tilgang til selve tabellen/viewet
resource "google_bigquery_table_iam_member" "metabase" {
  project    = "${project}"
  dataset_id = "${dataset}"
  table_id   = "${table}"
  role       = "roles/bigquery.dataViewer"
  member     = "serviceAccount:${saEmail}"
}

# Metadatatilgang på dataset-nivå (for at Metabase skal kunne se at tabellen/viewet eksisterer)
resource "google_bigquery_dataset_iam_member" "metabase_metadata" {
  project    = "${project}"
  dataset_id = "${dataset}"
  role       = "roles/bigquery.metadataViewer"
  member     = "serviceAccount:${saEmail}"
}`

  const terraformViewSnippet =
    `# Autoriser viewet på hvert kilde-dataset det leser fra:
resource "google_bigquery_dataset_access" "authorized_view" {
  project    = "<kilde-prosjekt>"
  dataset_id = "<kilde-dataset>"

  view {
    project_id = "${project}"
    dataset_id = "${dataset}"
    table_id   = "${table}"
  }
}`

  const sqlSnippet =
    `-- Tilgang til selve tabellen/viewet
GRANT \`roles/bigquery.dataViewer\`
ON TABLE \`${project}.${dataset}.${table}\`
TO 'serviceAccount:${saEmail}';

-- Metadatatilgang på dataset-nivå (for at Metabase skal kunne se at tabellen/viewet eksisterer)
GRANT \`roles/bigquery.metadataViewer\`
ON SCHEMA \`${project}.${dataset}\`
TO 'serviceAccount:${saEmail}';`

  return (
    <div>
      <ReadMore header="Metabase-tilgang i kode">
      <div className="flex flex-col gap-4 mt-1">
        <BodyShort size="small" className="text-ax-text-neutral-subtle">
          Markedsplassen setter opp Metabase-tilgangen automatisk. Hvis dbt, Terraform eller andre
          verktøy som styrer IAM overskriver disse tilgangene, kan du legge dem eksplisitt inn i
          koden din slik at de ikke forsvinner.
        </BodyShort>

        <div className="flex flex-col gap-1">
          <Label size="small">Metabase-servicebruker</Label>
          <div className="flex items-center gap-2">
            <BodyShort size="small" className="font-mono bg-ax-bg-neutral-moderate rounded px-3 py-1">
              {saEmail}
            </BodyShort>
            <Copy text={saEmail} />
          </div>
        </div>

        {isView && (
          <Alert variant="warning" size="small">
            Dette er et <strong>view</strong>. I tillegg til grantene under må viewet autoriseres
            på hvert dataset det leser fra. Se Terraform-fanen for snippet, eller gjør det i
            BigQuery-konsollen under «Authorized views».
          </Alert>
        )}

        <Tabs defaultValue="dbt" size="small">
          <Tabs.List>
            <Tabs.Tab value="dbt" label="dbt" />
            <Tabs.Tab value="terraform" label="Terraform" />
            <Tabs.Tab value="sql" label="BigQuery SQL" />
          </Tabs.List>

          <Tabs.Panel value="dbt" className="flex flex-col gap-3 pt-3">
            <SnippetBlock
              label="grants i modell-config (schema.yml eller øverst i .sql-fil)"
              code={dbtSnippet}
            />
            <BodyShort size="small" className="text-ax-text-neutral-subtle">
              💡 Bruker du dbt i både dev og prod? Bruk Jinja:{' '}
              <code className="font-mono">
                {"{{ [sa_dev] if target.name == 'dev' else [sa_prod] }}"}
              </code>
            </BodyShort>
            <Alert variant="info" size="small">
              dbt <code>grants</code>-config setter kun tilgang på tabell/view-nivå. Hvis dbt
              også overskriver IAM på dataset-nivå, må <strong>metadataViewer</strong> legges
              til via Terraform eller BigQuery SQL (se de andre fanene).
            </Alert>
          </Tabs.Panel>

          <Tabs.Panel value="terraform" className="flex flex-col gap-3 pt-3">
            <SnippetBlock label="IAM-ressurser" code={terraformSnippet} />
            {isView && (
              <SnippetBlock
                label="Autoriser view på kilde-dataset(s) — gjenta for hvert dataset viewet leser fra"
                code={terraformViewSnippet}
              />
            )}
          </Tabs.Panel>

          <Tabs.Panel value="sql" className="flex flex-col gap-3 pt-3">
            <SnippetBlock label="GRANT-statements" code={sqlSnippet} />
            {isView && (
              <Alert variant="info" size="small">
                Autorisering av views støttes ikke via SQL — bruk Terraform eller
                BigQuery-konsollen under «Authorized views» på kilde-datasettet.
              </Alert>
            )}
          </Tabs.Panel>
        </Tabs>
      </div>
    </ReadMore>
    </div>
  )
}

export default MetabaseGrantSnippet
