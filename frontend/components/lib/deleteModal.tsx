import * as React from 'react'
import { Modal, Button, Alert, BodyShort } from '@navikt/ds-react'

interface DeleteModalProps {
	open: boolean
	onCancel: () => void
	onConfirm: () => void
	name: string
	error: string
	resource: string
}
export const DeleteModal = ({
	open,
	onCancel,
	onConfirm,
	name,
	error,
	resource,
}: DeleteModalProps) => {
	const action = resource === "metabase-dashboard" ? "Sett som privat" : "Slett"
	return (
		<Modal open={open} onClose={onCancel} header={{ heading: action }}>
			<Modal.Body className="flex flex-col gap-4">
				<ConfirmText name={name} resource={resource} />
				<div className="flex flex-row gap-3">
					<Button variant="secondary" onClick={onCancel}>
						Avbryt
					</Button>
					<Button onClick={onConfirm}>{action}</Button>
				</div>
				{error && <Alert variant={'error'}>{error}</Alert>}
			</Modal.Body>
		</Modal>
	)
}
export default DeleteModal

const ConfirmText = ({name, resource}: {name: string, resource: string}) =>
  resource === "metabase-dashboard" ? (
    <>
      <BodyShort>
        Er du sikker på at du vil gjøre <strong>{name}</strong> privat?
      </BodyShort>
      <BodyShort>
        Dashboardet slettes ikke. Brukere som har tilgang vil fortsatt finne det i Metabase.
      </BodyShort>
    </>
  ) : (
    <BodyShort>
      Er du sikker på at du vil slette <strong>{name}</strong>?
    </BodyShort>
  )

