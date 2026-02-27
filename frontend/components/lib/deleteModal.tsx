import * as React from 'react'
import { useState } from 'react'
import { Modal, Button, Alert, BodyShort, Checkbox } from '@navikt/ds-react'

interface DeleteModalProps {
	open: boolean
	onCancel: () => void
	onConfirm: () => void
	name: string
	error: string
	resource: string
	warning?: string
	confirmText?: string
}
export const DeleteModal = ({
	open,
	onCancel,
	onConfirm,
	name,
	error,
	resource,
	warning,
	confirmText,
}: DeleteModalProps) => {
	const action = resource === "metabase-dashboard" ? "Fjern public lenke" : "Slett"
	const [confirmed, setConfirmed] = useState(false)
	const [loading, setLoading] = useState(false)

	const handleCancel = () => {
		setConfirmed(false)
		setLoading(false)
		onCancel()
	}

	const handleConfirm = () => {
		setConfirmed(false)
		setLoading(true)
		onConfirm()
	}

	return (
		<Modal open={open} onClose={handleCancel} header={{ heading: action }}>
			<Modal.Body className="flex flex-col gap-4">
				<ConfirmText name={name} resource={resource} />
				{warning && <Alert variant="warning">{warning}</Alert>}
				{confirmText && (
					<Checkbox checked={confirmed} onClick={() => setConfirmed(!confirmed)}>
						{confirmText}
					</Checkbox>
				)}
				<div className="flex flex-row gap-3">
					<Button variant="secondary" onClick={handleCancel}>
						Avbryt
					</Button>
					<Button onClick={handleConfirm} disabled={!!confirmText && !confirmed} loading={loading}>{action}</Button>
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
        Er du sikker på at du vil fjerne public lenke til dashboardet <strong>{name}</strong>?
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

