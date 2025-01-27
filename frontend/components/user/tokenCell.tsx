import { Button, Heading, Modal, Table } from "@navikt/ds-react"
import { useState } from "react"
import { ArrowsCirclepathIcon, EyeFillIcon, EyeObfuscatedFillIcon } from '@navikt/aksel-icons';
import { useRouter } from "next/navigation";
import { updateTeamToken } from "../../lib/rest/userData";

const TokenCell = ({token, team}: {token: string, team: string}) => {
    const [hidden, setHidden] = useState(true)
    const [showRotateModal, setShowRotateModal] = useState(false)
    const router = useRouter()

    const rotateTeamToken = (team: string) => {
          updateTeamToken(team)
          .then(() => {
            setShowRotateModal(false)
            router.refresh()
            })
          .catch(err => {
            //TODO: show error
            console.log("err", err)
          })
    }

    return (
        <>
        <Modal
                open={showRotateModal}
                aria-label="Roter nada token"
                onClose={() => setShowRotateModal(false)}
                className="max-w-full md:max-w-3xl px-8 h-[20rem]"
              >
                <Modal.Body className="h-full">
                  <div className="flex flex-col gap-8">
                    <Heading level="1" size="medium">
                    Er du sikker på at du vil rotere tokenet for team {team}?
                    </Heading>
                    <div>
                        Dette vil innebære at du må oppdatere til nytt token i all kode der dette brukes for å autentisere seg mot datamarkedsplassen.
                    </div>
                    <div className="flex flex-row gap-4">
                      <Button
                        onClick={() => rotateTeamToken(team)}
                        variant="primary"
                        size="small"
                      >
                        Roter
                      </Button>
                      <Button
                        onClick={() => setShowRotateModal(false)}
                        variant="secondary"
                        size="small"
                      >
                        Avbryt
                      </Button>
                    </div>
                  </div>
                </Modal.Body>
              </Modal>
        <Table.DataCell className="flex flex-row gap-2 items-center w-full">
            <div>
                {hidden 
                    ? (<EyeFillIcon className="cursor-pointer" onClick={() => setHidden(!hidden)} />) 
                    : (<EyeObfuscatedFillIcon className="cursor-pointer" onClick={() => setHidden(!hidden)} />)}
            </div>
            <span>{hidden 
                ? "*********"
                : token
            }</span>
            <ArrowsCirclepathIcon className="cursor-pointer" title="roter token" fontSize="1.2rem" onClick={() => setShowRotateModal(true)}/>
        </Table.DataCell>
        </>
    )
}

export default TokenCell