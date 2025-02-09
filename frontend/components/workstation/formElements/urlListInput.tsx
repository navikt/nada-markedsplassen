import { Label, Link, Textarea } from "@navikt/ds-react";
import { ExternalLinkIcon } from "@navikt/aksel-icons";
import { useImperativeHandle, useState } from 'react'

export interface UrlListInputProps {
    urlList: string[];
    ref: React.Ref<{ getUrls: () => string[] }>;
}

export const UrlListInput = (props: UrlListInputProps) => {
  const [urlList, setUrlList] = useState<string[]>(props.urlList);

  useImperativeHandle(props.ref, () => ({
    getUrls: () => {
      return urlList;
    },
  }))

  function handleChange(event: React.ChangeEvent<HTMLTextAreaElement>) {
        const urlList = event.target.value.split("\n").filter((url) => url !== "")
        setUrlList(urlList)
    }

    return (
        <div className="flex gap-2 flex-col">
            <Label>Oppgi hvilke internett-URLer du vil åpne mot</Label>
            <p className="pt-0">
                Du kan legge til opptil 2500 oppføringer i en URL-liste. Hver oppføring må stå på en egen linje uten
                mellomrom eller skilletegn. Oppføringer kan være kun domenenavn (som matcher alle stier) eller inkludere
                en sti-komponent.{' '}
                <Link target="_blank" href="https://cloud.google.com/secure-web-proxy/docs/url-list-syntax-reference">
                    Les mer om syntax her <ExternalLinkIcon/>
                </Link>
            </p>
            <Textarea
                defaultValue={props.urlList.join("\n")}
                size="small"
                maxRows={2500}
                hideLabel
                label="Hvilke URLer vil du åpne mot"
                resize
                onChange={handleChange}
            />
        </div>
    );
}

export default UrlListInput;
