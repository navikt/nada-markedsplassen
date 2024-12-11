import {Textarea, Label, Link} from "@navikt/ds-react";
import {ExternalLink} from "@navikt/ds-icons";
import {useEffect} from "react";
import {useWorkstationMine} from "../queries";

export interface UrlListInputProps {
    initialUrlList?: string[];
    onUrlListChange: (urlList: string[]) => void;
}

export const UrlListInput = (props: UrlListInputProps) => {
    const workstation = useWorkstationMine()

    const urlList = props.initialUrlList ?? workstation.data?.urlAllowList ?? []

    useEffect(() => {
        props.onUrlListChange(urlList)
    }, [urlList])

    if (workstation.isLoading) {
        return <Textarea label="Hvilke URL-er vil du åpne mot" defaultValue="Laster..." size="small" maxRows={2500}
                         readOnly resize/>
    }

    function handleChange(event: React.ChangeEvent<HTMLTextAreaElement>) {
        const urlList = event.target.value.split("\n").filter((url) => url !== "")
        props.onUrlListChange(urlList)
    }

    return (
        <div className="flex gap-2 flex-col">
            <Label>Oppgi hvilke internett-URL-er du vil åpne mot</Label>
            <p className="pt-0">
                Du kan legge til opptil 2500 oppføringer i en URL-liste. Hver oppføring må stå på en egen linje uten
                mellomrom eller skilletegn. Oppføringer kan være kun domenenavn (som matcher alle stier) eller inkludere
                en sti-komponent.{' '}
                <Link target="_blank" href="https://cloud.google.com/secure-web-proxy/docs/url-list-syntax-reference">
                    Les mer om syntax her <ExternalLink/>
                </Link>
            </p>
            <Textarea
                defaultValue={urlList ? urlList.length > 0 ? urlList.join("\n") : "" : ""}
                size="small"
                maxRows={2500}
                hideLabel
                label="Hvilke URL-er vil du åpne mot"
                resize
                onChange={handleChange}
            />
        </div>
    );
}

export default UrlListInput;
