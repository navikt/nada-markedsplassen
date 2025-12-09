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
        {page === "loading" ? <div>Laster knast informasjon... <Loader /></div>
            : <div>
                {page === "create" ? <CreateKnastForm options={knastOptions.data} />
                    : <ManageKnastPage />}
                <p className="italic text-sm text-gray-600">Knast UI er i beta, vennligst skriv i Slack-kanalen #dataplattform-arbeidsflater for tilbakemelding.</p>
            </div>
        }
    </div>
    )
}

export default Knast
