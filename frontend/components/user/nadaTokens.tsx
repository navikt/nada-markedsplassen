import React, {useState} from 'react'
import {Table} from "@navikt/ds-react";
import TokenCell, {RotateTokenModal} from './tokenCell';

/** NadaToken contains the team token of the corresponding team for updating data stories */
export type NadaToken = {
    __typename?: 'NadaToken';
    /** name of team */
    team: string;
    /** nada token for the team */
    token: string;
  };

const NadaTokensForUser = ({ nadaTokens }: { nadaTokens: Array<NadaToken> }) => {
    const [rotateTeam, setRotateTeam] = useState<string | null>(null)

    return (
        <>
            {rotateTeam && (
                <RotateTokenModal
                    team={rotateTeam}
                    open={true}
                    onClose={() => setRotateTeam(null)}
                />
            )}
            <Table zebraStripes>
                <Table.Header>
                    <Table.Row>
                        <Table.HeaderCell scope="col">Team</Table.HeaderCell>
                        <Table.HeaderCell scope="col">Token</Table.HeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                {nadaTokens.map((token, i) => 
                    <Table.Row key={i}>
                        <Table.DataCell className="w-96" scope="row">{token.team}</Table.DataCell>
                        <TokenCell token={token.token} team={token.team} onRotate={() => setRotateTeam(token.team)}/>
                    </Table.Row>
                )}
                </Table.Body>
            </Table>
        </>
    )
}

export default NadaTokensForUser
