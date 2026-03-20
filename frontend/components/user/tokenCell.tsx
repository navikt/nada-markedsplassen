import {Button, CopyButton, Heading, Modal, Table} from "@navikt/ds-react"
import React, {useState} from "react"
import {ArrowsCirclepathIcon, EyeFillIcon, EyeObfuscatedFillIcon} from '@navikt/aksel-icons';
import {useRouter} from "next/navigation";
import {updateTeamToken} from "../../lib/rest/userData";

export const RotateTokenModal = ({team, open, onClose}: { team: string, open: boolean, onClose: () => void }) => {
    const router = useRouter()

    const rotateTeamToken = (team: string) => {
        updateTeamToken(team)
            .then(() => {
                onClose()
                router.refresh()
            })
            .catch(err => {
                //TODO: show error
                console.log("err", err)
            })
    }

    return (
        <Modal
            open={open}
            aria-label="Roter nada token"
            onClose={onClose}
            className="max-w-full ax-md:max-w-3xl px-8 h-[20rem]"
        >
            <Modal.Body className="h-full">
                <div className="flex flex-col gap-8">
                    <Heading level="1" size="medium">
                        Er du sikker på at du vil rotere tokenet for team {team}?
                    </Heading>
                    <div>
                        Dette vil innebære at du må oppdatere til nytt token i all kode der dette brukes for å
                        autentisere seg mot datamarkedsplassen.
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
                            onClick={onClose}
                            variant="secondary"
                            size="small"
                        >
                            Avbryt
                        </Button>
                    </div>
                </div>
            </Modal.Body>
        </Modal>
    )
}

const TokenCell = ({token, team, onRotate}: { token: string, team: string, onRotate: () => void }) => {
    const [hidden, setHidden] = useState(true)

    return (
        <Table.DataCell className="flex flex-row gap-2 items-center w-full">
            <span>{hidden ? "*********" : token}</span>
            <div>
                {hidden
                    ? (<EyeFillIcon className="cursor-pointer" onClick={() => setHidden(!hidden)}/>)
                    : (<EyeObfuscatedFillIcon className="cursor-pointer" onClick={() => setHidden(!hidden)}/>)}
            </div>
            <CopyButton copyText={token}/>
            <ArrowsCirclepathIcon className="cursor-pointer" title="roter token" fontSize="1.2rem"
                                  onClick={onRotate}/>
        </Table.DataCell>
    )
}

export default TokenCell