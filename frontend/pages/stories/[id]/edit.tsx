import { useRouter } from "next/router"
import { EditStoryMetadataForm } from "../../../components/stories/editStoryMetadata";
import LoaderSpinner from "../../../components/lib/spinner";
import { useEffect, useState } from "react";
import ErrorMessage from "../../../components/lib/error";
import { fetchStoryMetadata } from "../../../lib/rest/stories";

export const useGetStoryMetadata = (id: string)=>{
    const [storyMetadata, setStoryMetadata] = useState<any>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)


    useEffect(()=>{
        if(!id) return
        fetchStoryMetadata(id)
        .then((story)=>
        {
            setError(null)
            setStoryMetadata(story)
        })
        .catch((err)=>{
            setError(err)
            setStoryMetadata(null)            
        }).finally(()=>{
            setLoading(false)
        })
    }, [id])

    return {storyMetadata, loading, error}
}


const EditStoryPage = ()=>{
    const router = useRouter()
    const id = router.query.id;
    const data = useGetStoryMetadata(id as string)

    if (data.error) return <ErrorMessage error={data.error} />
    if (data.loading || !data.storyMetadata)
      return <LoaderSpinner />

    const story = data.storyMetadata

    return <div>
        <EditStoryMetadataForm 
            id={id as string} 
            name={story.name} 
            description={story.description} 
            keywords={story.keywords} 
            teamkatalogenURL={story.teamkatalogenURL || ""} 
            group={story.group} />
    </div>
}

export default EditStoryPage;