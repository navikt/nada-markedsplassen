import { useRouter } from "next/router"
import { useEffect } from "react"

//the page proxy for umami analytics
const StoryProxy = ()=>{
    const router = useRouter()

    useEffect(()=>{
        const handleRouteChange = (url: string) => {
            if (url === '/') {
                if (typeof window !== 'undefined' && window.umami) {
                    window.umami.track('view-story', {id: router.query.id})
                }
                if(router.query.id){
                    router.push(`/quarto/${router.query.id}`)
                }
            }
        }

        router.events.on('routeChangeComplete', handleRouteChange)

        return () => {
            router.events.off('routeChangeComplete', handleRouteChange)
        }
    }, [router.query.id])
    
    return <>
        <p>Omdirigere...</p>
    </>;
}

export default StoryProxy