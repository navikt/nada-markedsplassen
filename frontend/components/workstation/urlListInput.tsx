import { Textarea, Label, Link } from "@navikt/ds-react";
import { ExternalLink } from "@navikt/ds-icons";

interface UrlListInputProps {
    urlList: string[];
    defaultUrlList: string[];
    onUrlListUpdate: (event: any) => void;
}

const UrlListInput: React.FC<UrlListInputProps> = ({ urlList, defaultUrlList, onUrlListUpdate }) => {
    return (
        <div className="flex gap-2 flex-col">
            <Label>Oppgi hvilke internett-URL-er du vil åpne mot</Label>
            <p className="pt-0">
                Du kan legge til opptil 2500 oppføringer i en URL-liste. Hver oppføring må stå på en egen linje uten mellomrom eller skilletegn. Oppføringer kan være kun domenenavn (som matcher alle stier) eller inkludere en sti-komponent.
                <Link target="_blank" href="https://cloud.google.com/secure-web-proxy/docs/url-list-syntax-reference">
                    Les mer om syntax her <ExternalLink />
                </Link>
            </p>
            <Textarea
                onChange={onUrlListUpdate}
                defaultValue={urlList.length > 0 ? urlList.join("\n") : defaultUrlList.join("\n")}
                size="medium"
                maxRows={2500}
                hideLabel
                label="Hvilke URL-er vil du åpne mot"
                resize
            />
        </div>
    );
};

export default UrlListInput;
