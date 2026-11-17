import { Loader } from "@navikt/ds-react";
import { useRouter } from "next/router";
import { EditMetabaseResource } from "../../components/metabaseDashboards/EditMetabaseResource";
import { useGetMetabaseDashboardUrl } from "../../lib/rest/metabaseDashboards";

const EditMetabaseDashboardPage = () => {
    const router = useRouter();
    const id = router.query.id;
    const metabaseDashboardQuery = useGetMetabaseDashboardUrl(id as string);

    if (metabaseDashboardQuery.error) {
        return <div>{metabaseDashboardQuery.error.message}</div>;
    }

    if (metabaseDashboardQuery.isLoading || !metabaseDashboardQuery.data) {
        return <Loader/>;
    }

    const metabaseDashboard = metabaseDashboardQuery.data;

    return (
        <div>
            <EditMetabaseResource
                resourceType="dashboard"
                id={id as string}
                name={metabaseDashboard.name}
                description={metabaseDashboard.description}
                keywords={metabaseDashboard.keywords}
                teamkatalogenURL={metabaseDashboard.teamkatalogenURL || ""}
                group={metabaseDashboard.group}
                link={metabaseDashboard.link}
                teamID={metabaseDashboard.teamID || ""}
            />
        </div>
    );
};

export default EditMetabaseDashboardPage;
