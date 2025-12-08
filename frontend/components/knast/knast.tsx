import { useWorkstationExists, useWorkstationJobs, useWorkstationOptions } from "./queries";
import { Loader } from "@navikt/ds-react";
import CreateKnastForm from "./createKnastForm";
import { ManageKnastPage } from "./manageKnast";

const Knast = () => {
    const workstationExists = useWorkstationExists()
    const workstationJobs = useWorkstationJobs()
    const knastOptions = useWorkstationOptions()

    const page = workstationExists.isLoading || workstationJobs.isLoading
        ? "loading"
        : !workstationExists.data
            ? "create"
            : "manage"

    return (<div>
        <p className="italic text-sm text-gray-600">Knast UI er i beta, vennligst skriv i Slack-kanalen #dataplattform-arbeidsflater for tilbakemelding.</p>
        {page === "loading" ? <div>Laster knast informasjon... <Loader /></div>
            : page === "create" ? <CreateKnastForm options={knastOptions.data} />
                : <ManageKnastPage />}
    </div>
    )
}

export default Knast
