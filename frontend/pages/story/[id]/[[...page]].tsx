import { useRouter } from "next/router"
import { useEffect } from "react"

//the page proxy for umami analytics
const StoryProxy = ()=>{
    const router = useRouter()
    console.log(router.query)

    useEffect(()=>{
        if(router.query.id){
            router.push(`/quarto/${router.query.id}`)
        }
    })
    return <>
        <p>Omdirigere...</p>
    </>;
}

export default StoryProxy