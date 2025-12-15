import { useState } from 'react'
import { parseISO, format } from 'date-fns'
import {
  Alert,
  BodyShort,
  Button,
  Heading,
  Link,
  Loader,
  Modal,
  Table,
  Textarea,
} from '@navikt/ds-react'
import { ExternalLinkIcon, DatabaseIcon, AreaChartIcon } from '@navikt/aksel-icons'
import { nb } from 'date-fns/locale'
import { useGetDataset } from '../../../lib/rest/dataproducts'
import { apporveAccessRequest, denyAccessRequest, revokeAllUsersMetabaseAccess, revokeDatasetAccess, revokeRestrictedMetabaseAccess, useFetchAccessRequestsForDataset } from '../../../lib/rest/access'
import { Access, DatasetAccess as dtoDatasetAccess } from '../../../lib/rest/generatedDto'
import ErrorStripe from '../../lib/errorStripe'

const humanizeDateAccessForm = (
  isoDate: string,
  dateFormat = 'dd. MMMM yyyy'
) => {
  try {
    const parsed = parseISO(isoDate)
    return (
      <time
        dateTime={isoDate}
        title={format(parsed, 'dd. MMMM yyyy HH:mm:ii', { locale: nb })}
      >
        {format(parsed, dateFormat, { locale: nb })}
      </time>
    )
  } catch (e) {
    return <></>
  }
}

const lookupUserAccessesAcrossPlatforms = (datasetAccess: any, subject: string) => {
  const activeAccesses: Access[] = []

  datasetAccess.forEach((a: dtoDatasetAccess) => {
    if (a.subject === subject) {
      a.active.forEach((access: Access | undefined) => {
        if (access !== undefined) activeAccesses.push(access)
      })
    }
  })

  return activeAccesses
}

interface AccessListProps {
  id: string
}

interface AccessModalProps {
  subject: string
  datasetName: string
  action: () => Promise<void>
}

interface AccessRequestModalProps {
  requestID: string
  user?: string
}

export const AccessRequestModal = ({
  requestID,
  user,
}: AccessRequestModalProps) => {
  const [submitted, setSubmitted] = useState(false)
  const [openDeny, setOpenDeny] = useState(false)
  const [openApprove, setOpenApprove] = useState(false)
  const [errorApprove, setErrorApprove] = useState<string | undefined>(undefined)
  const [errorDeny, setErrorDeny] = useState<string | undefined>(undefined)
  const [reason, setReason] = useState<string>('')
  const approve = async (requestID: string) =>
    apporveAccessRequest(requestID).then(() => {
      setOpenApprove(false)
      setErrorApprove(undefined)
      window.location.reload();
    }
    ).catch((e: any) => {
      setErrorApprove(e.message)
    })
  const deny = async (requestID: string, reason?: string) => denyAccessRequest(requestID, reason || '')
    .then(() => {
      setOpenDeny(false)
      setErrorDeny(undefined)
      window.location.reload();
    }).catch((e: any) => {
      setErrorDeny(e.message)
    })

  const cancelApprove = () => {
    setOpenApprove(false)
    setErrorApprove(undefined)
  }

  const cancelDeny = () => {
    setOpenDeny(false)
    setErrorDeny(undefined)
  }

  return (
    <>
      <Modal
        open={openApprove}
        aria-label="Godkjenn søknad"
        onClose={() => setOpenApprove(false)}
        className='w-full md:w-240 px-8 h-52'
      >
        <Modal.Body className='h-full'>
          <div className='flex flex-col justify-center items-center'>
            <Heading level="1" size="medium">
              Godkjenn søknad
            </Heading>
            <p className='mt-4 mb-4'>Gi tilgang til datasett{user ? ` til ${user}` : ''}? </p>
            <div className="flex flex-col gap-4 items-center">
              <div className="flex flex-row gap-4">
                <Button
                  onClick={cancelApprove}
                  variant="secondary"
                  size="small"
                >
                  Avbryt
                </Button>
                <Button
                  onClick={() => {
                    setSubmitted(true)
                    approve(requestID)
                  }}
                  variant="primary"
                  size="small"
                  disabled={submitted}
                  loading={submitted}
                >
                  Godkjenn
                </Button>
              </div>
              {errorApprove && <div className='text-red-600'>{errorApprove}</div>}
              {submitted && !errorApprove && <div>Vennligst vent...<Loader size="small" /></div>}
            </div>
          </div>
        </Modal.Body>
      </Modal>

      <Modal
        open={openDeny}
        aria-label="Avslå søknad"
        onClose={() => setOpenDeny(false)}
        className="max-w-full md:max-w-3xl px-8 h-96"
      >
        <Modal.Body className="h-full">
          <div className="flex flex-col items-center gap-8">
            <Heading level="1" size="medium">
              Avslå søknad
            </Heading>
            <Textarea label="Begrunnelse" value={reason} onChange={event => setReason(event.target.value)} />
            <div className="flex flex-col gap-4">
              <div className="flex flex-row gap-4">
                <Button
                  onClick={cancelDeny}
                  variant="secondary"
                  size="small"
                >
                  Avbryt
                </Button>
                <Button
                  onClick={() => {
                    setSubmitted(true)
                    deny(requestID, reason)
                  }
                  }
                  variant="primary"
                  size="small"
                  disabled={submitted}
                >
                  Avslå
                </Button>
              </div>
              {errorDeny && <div className='text-red-600'>{errorDeny}</div>}
              {submitted && !errorDeny && <div>Vennligst vent...<Loader size="small" /></div>}
            </div>
          </div>
        </Modal.Body>
      </Modal>
      <div className="flex flex-row flex-nowrap gap-4 justify-end">
        <Button
          onClick={() => setOpenApprove(true)}
          variant="secondary"
          size="small"
        >
          Godkjenn
        </Button>
        <Button onClick={() => setOpenDeny(true)} variant="secondary" size="small">
          Avslå
        </Button>
      </div>
    </>
  )
}

export const AccessModal = ({ subject, datasetName, action }: AccessModalProps) => {
  const [open, setOpen] = useState(false)
  const [removingAccess, setRemovingAccess] = useState(false)
  const [error, setError] = useState<string | undefined>()

  const onSubmit = async () => {
    try {
      setRemovingAccess(true)
      await action()
      setOpen(false)
      setError(undefined)
    } catch {
      setError("Noe gikk galt, prøv igjen...")
      setRemovingAccess(false)
    }
  }

  console.log(error)

  return (
    <>
      <Modal
        open={open}
        onClose={() => {
          setOpen(false)
          setError(undefined)
        }}
        header={{
          heading: "Fjern tilgang",
        }}
        className='w-fit'
      >
        <Modal.Body>
          <BodyShort className='mb-1'>
            Du fjerner nå tilgang til datasettet{' '}
            <strong>{datasetName}</strong> for{' '} 
            <span className="whitespace-nowrap font-bold">
              {subject.split(':')[1]}
            </span>
          </BodyShort>
          <BodyShort>
            Er du sikker?
          </BodyShort>
          {error && <Alert variant='error' size='small' className='mt-3'>{error}</Alert>}
        </Modal.Body>
        <Modal.Footer>
          <Button
            onClick={onSubmit}
            variant="primary"
            size="small"
            loading={removingAccess}
          >
            Fjern
          </Button>
          <Button
            onClick={() => setOpen(false)}
            variant="secondary"
            size="small"
          >
            Avbryt
          </Button>
        </Modal.Footer>
      </Modal>

      <Button
        onClick={() => setOpen(true)}
        className="flex-nowrap"
        variant="secondary"
        size="small"
      >
        Fjern tilgang
      </Button>
    </>
  )
}

const DatasetAccess = ({ id }: AccessListProps) => {
  const [formError, setFormError] = useState('')
  const fetchAccessRequestsForDataset = useFetchAccessRequestsForDataset(id)

  const getDataset = useGetDataset(id)

  if (fetchAccessRequestsForDataset.error)
    return <ErrorStripe error={fetchAccessRequestsForDataset.error} />

  const datasetAccessRequests = fetchAccessRequestsForDataset.isLoading ||
    !fetchAccessRequestsForDataset.data
    ? undefined
    : fetchAccessRequestsForDataset.data

  if (getDataset.error)
    return <ErrorStripe error={getDataset.error} />

  const access = getDataset.isLoading ||
    !getDataset?.data?.access ? [] :
    getDataset.data.access

  const removeAccess = async (subject: string) => {
    const accessesForUser = lookupUserAccessesAcrossPlatforms(access, subject)
    try {
      for (const a of accessesForUser) {
        try {
          if (a.platform === 'bigquery') {
            await revokeDatasetAccess(a.id)
          } else if (a.platform === 'metabase') {
            if (a.subject === 'group:all-users@nav.no') {
              await revokeAllUsersMetabaseAccess(a.id)
            } else {
              await revokeRestrictedMetabaseAccess(a.id)
            }
          }
        } catch (e: any) {
          setFormError(e.message)
        }
      }
    } catch (e: any) {
      setFormError(e.message)
    } finally {
      if (!formError) {
        window.location.reload();
      }
    }
  }

  return (
    <div className="flex flex-col gap-8 w-full 2xl:w-240">
      {formError && <Alert variant={'error'}>{formError}</Alert>}
      <div>
        <Heading level="2" size="small">
          Tilgangssøknader
        </Heading>
        <div className="mb-3 w-[91vw] md:w-auto overflow-auto">
          {datasetAccessRequests?.accessRequests.length ? (
            <Table>
              <Table.Header>
                <Table.Row>
                  <Table.HeaderCell>Bruker/gruppe</Table.HeaderCell>
                  <Table.HeaderCell>Brukertype</Table.HeaderCell>
                  <Table.HeaderCell>Tilgang</Table.HeaderCell>
                  <Table.HeaderCell>Plattform</Table.HeaderCell>
                  <Table.HeaderCell />
                  <Table.HeaderCell />
                </Table.Row>
              </Table.Header>
              {datasetAccessRequests.accessRequests.map((r, i) => (
                <>
                  <Table.Row
                    className={i % 2 === 0 ? 'bg-[#f7f7f7]' : ''}
                    key={i + '-request'}
                  >
                    <Table.DataCell className="w-72">{r.subject}</Table.DataCell>
                    <Table.DataCell className="w-36">
                      {r.subjectType}
                    </Table.DataCell>
                    <Table.DataCell className="w-48">
                      {r.expires
                        ? humanizeDateAccessForm(r.expires)
                        : 'For alltid'}
                    </Table.DataCell>
                    <Table.DataCell className="w-48">
                      <div className='flex flex-row gap-1 items-center'>
                        {r.platform} {r.platform === 'bigquery' ? <DatabaseIcon /> : r.platform === 'metabase' ? <AreaChartIcon /> : null}
                      </div>
                    </Table.DataCell>
                    <Table.DataCell className="w-48">
                      {r.polly?.url ? (
                        <Link target="_blank" rel="norefferer" href={r.polly?.url}>
                          Åpne behandling
                          <ExternalLinkIcon />
                        </Link>
                      ) : (
                        'Ingen behandling'
                      )}
                    </Table.DataCell>
                    <Table.DataCell className="w-[150px]" align="right">
                      <AccessRequestModal
                        requestID={r.id}
                        user={r.subject}
                      />
                    </Table.DataCell>
                  </Table.Row>
                </>
              ))}
            </Table>
          ) : (
            'Ingen tilgangssøknader'
          )}
        </div>
      </div>
      <div>
        <Heading level="2" size="small">
          Aktive tilganger
        </Heading>
        <div className="mb-3 w-[91vw] md:w-auto overflow-auto">

          {access.flatMap(a => a?.active).length > 0 ? (
            <Table>
              <Table.Header>
                <Table.Row className="border-none border-transparent">
                  <Table.HeaderCell>Bruker/gruppe</Table.HeaderCell>
                  <Table.HeaderCell>Brukertype</Table.HeaderCell>
                  <Table.HeaderCell>Tilgang</Table.HeaderCell>
                  <Table.HeaderCell>Plattformer</Table.HeaderCell>
                  <Table.HeaderCell />
                  <Table.HeaderCell />
                </Table.Row>
              </Table.Header>
              {access.filter(a => { return a !== undefined }).filter(a => a.active.length > 0).map((a, i) => (
                <>
                  <Table.Row
                    className={i % 2 === 0 ? 'bg-[#f7f7f7]' : ''}
                    key={i + '-access'}
                  >
                    <Table.DataCell className="w-72">{a?.subject.split(':')[1]}</Table.DataCell>
                    <Table.DataCell className="w-36">
                      {a.subject.split(':')[0]}
                    </Table.DataCell>
                    <Table.DataCell className="w-48">
                      {a.active[0]?.expires
                        ? humanizeDateAccessForm(a.active[0].expires)
                        : 'For alltid'}
                    </Table.DataCell>
                    <Table.DataCell className="w-48">
                      <div className='flex flex-row gap-1 items-center'>
                        {a.active.map(acc => acc?.platform).join(', ')}
                      </div>
                    </Table.DataCell>
                    <Table.DataCell className="w-48">
                      {a.active[0]?.accessRequest?.polly !== undefined && a.active[0].accessRequest?.polly?.url !== "" ? (
                        <Link
                          target="_blank"
                          rel="norefferer"
                          href={a.active[0].accessRequest?.polly?.url}
                        >
                          Åpne behandling
                          <ExternalLinkIcon />
                        </Link>
                      ) : (
                        'Ingen behandling'
                      )}
                    </Table.DataCell>
                    <Table.DataCell className="w-[207px]" align="left">
                      <AccessModal subject={a.subject} datasetName={getDataset.data?.name || ""} action={() => removeAccess(a.subject)} />
                    </Table.DataCell>
                  </Table.Row>
                </>
              )
              )}
            </Table>
          ) : (
            'Ingen aktive tilganger'
          )}
        </div>
      </div>
    </div>
  )
}

export default DatasetAccess
