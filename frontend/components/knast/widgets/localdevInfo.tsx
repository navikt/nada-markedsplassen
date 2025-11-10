import { BodyLong, CopyButton, Link, List, Modal } from "@navikt/ds-react"
import { WorkstationOutput } from "../../../lib/rest/generatedDto"

export const LocalDevInfo = ({ show, knastInfo, onClose }: {
    show: boolean
    knastInfo: WorkstationOutput
    onClose: () => void
}) => {
    console.log(knastInfo)
    return (
        <Modal open={show} onClose={onClose} width="medium" header={{ heading: 'Bruk av Knast via lokal IDE' }} closeOnBackdropClick>
            <Modal.Body>
                <BodyLong>
                    <List as="ol" title="Følgende må gjøres på lokal maskin for å koble til Knast:">
                        <List.Item title={'Logg inn i Google Cloud (kjøres lokalt)'}>
                            <div className="flex">
                                <code
                                    className="rounded-xs bg-surface-neutral-subtle font-mono text-sm font-semibold">gcloud
                                    auth login</code><CopyButton size="xsmall" copyText="gcloud auth login" />
                            </div>
                        </List.Item>
                        <List.Item title={'Opprette SSH-tunnel (kjøres lokalt)'}>
                            <div className="flex">
                                <code
                                    className="rounded-xs bg-surface-neutral-subtle px-1 py-05 font-mono text-sm font-semibold">
                                    gcloud workstations start-tcp-tunnel --cluster=knada
                                    --config={knastInfo.slug} --region=europe-north1 --project knada-gcp
                                    --local-host-port=:33649 {knastInfo.slug} 22</code>
                                <CopyButton size="xsmall"
                                    copyText={`gcloud workstations start-tcp-tunnel --cluster=knada --config=${knastInfo.slug} --region=europe-north1 --project knada-gcp --local-host-port=:33649 ${knastInfo.slug} 22`} />
                            </div>
                        </List.Item>
                        <List.Item title={'Opprette SSH-nøkkel (kjøres lokalt, hopp over om du allerede har gjort dette)'}>
                            <em>Sett et passord på SSH-nøkkelen. Du vil aldri bli bedt om å bytte dette.</em>
                            <div className="flex">
                                <code
                                    className="rounded-xs bg-surface-neutral-subtle font-mono text-sm font-semibold">{`ssh-keygen -t ed25519 -C "din_epost_email@nav.no"`}</code><CopyButton
                                        size="xsmall" copyText={`ssh-keygen -t ed25519 -C "din_epost_email@nav.no"`}></CopyButton>
                            </div>
                        </List.Item>
                        <List.Item title={'Få Knast til å stole på din SSH-nøkkel'}>
                            <div className="bg-red-50"><b> NB! Dette steget utføres <em>på Knasten</em>. Trykk &quot;Åpne din
                                Knast i nytt vindu&quot;</b></div>
                            <ul>
                                <li>Opprette directory <strong>~/.ssh/</strong> hvis det ikke allerede finnes</li>
                                <li>Opprett filen <strong>authorized_keys</strong> i <strong>~/.ssh/</strong></li>
                                <li>Lime inn innholdet fra public-delen av SSH-nøkkelen
                                    fra <strong>.ssh/id_ed25519.pub</strong> eller tilsvarende på lokal maskin inn
                                    i <strong>authorized_keys</strong> på Knast
                                </li>
                            </ul>

                        </List.Item>
                        <List.Item
                            title={'Legg til knast i ssh-configen (kjøres lokalt, hopp over om du allerede har gjort dette)'}>
                            <div className="flex">
                                <code
                                    className="rounded-xs bg-surface-neutral-subtle font-mono text-sm font-semibold">{`echo -e "\\nHost knast\\n\\tHostName localhost\\n\\tPort 33649\\n\\tUser user\\n\\tUserKnownHostsFile /dev/null\\n\\tStrictHostKeyChecking no">>~/.ssh/config`}</code><CopyButton
                                        size="xsmall"
                                        copyText={`echo -e "\\nHost knast\\n\\tHostName localhost\\n\\tPort 33649\\n\\tUser user\\n\\tUserKnownHostsFile /dev/null\\n\\tStrictHostKeyChecking no">>~/.ssh/config`}></CopyButton>
                            </div>
                        </List.Item>
                        {knastInfo.config?.image?.includes('intellij') || knastInfo.config?.image?.includes('pycharm') ?
                            (
                                <div>
                                    <List.Item
                                        title={'Sørg for at Remote Development Gateway pluginen er installert i IntelliJ/PyCharm'}>
                                        For installasjon av pluginen se <Link
                                            href="https://www.jetbrains.com/help/idea/jetbrains-gateway.html">her</Link>.
                                    </List.Item>
                                    <List.Item title={'Åpne Remote Development i IntelliJ/PyCharm (File | Remote Development...)'}>
                                        <ul>
                                            <li>Følg så denne <Link
                                                href="https://www.jetbrains.com/help/idea/remote-development-starting-page.html#connect_to_rd_ij">guiden</Link>
                                            </li>
                                        </ul>
                                    </List.Item>
                                </div>
                            ) : (
                                <div>
                                    <List.Item>
                                        Installere extension <code>Remote - SSH</code> i VS Code
                                    </List.Item>
                                    <List.Item title={'Åpne Command Palette i VS Code (⇧⌘P / Ctrl+Shift+P)'}>
                                        <ul>
                                            <li>Velg/Skriv inn: <code>Remote - SSH: Connect to host...</code></li>
                                            <li>Skriv inn: <strong>knast</strong></li>
                                        </ul>
                                    </List.Item>
                                    <p>Dette er også beskrevet med skjermbilder i <Link
                                        href="https://cloud.google.com/workstations/docs/develop-code-using-local-vscode-editor">dokumentasjonen
                                        til Google Workstations.</Link></p>
                                </div>
                            )}
                    </List>
                </BodyLong>
            </Modal.Body>
        </Modal>
    )
}
