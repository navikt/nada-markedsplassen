import { Loader } from "@navikt/ds-react";
import { useRouter } from "next/router";
import { EditMetabaseDashboard, EditMetabaseDashboardProps } from "../../components/metabaseDashboards/EditMetabaseDashboard";
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

    const formFields: EditMetabaseDashboardProps = {
        id: id as string,
        name: metabaseDashboard.name,
        description: metabaseDashboard.description,
        keywords: metabaseDashboard.keywords,
        teamkatalogenURL: metabaseDashboard.teamkatalogenURL || "",
        group: metabaseDashboard.group,
        link: metabaseDashboard.link,
        teamID: metabaseDashboard.teamID || "",
    };

    return (
        <div>
            <EditMetabaseDashboard {...formFields} />
        </div>
    );
};

export default EditMetabaseDashboardPage;