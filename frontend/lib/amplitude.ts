import amplitude from 'amplitude-js'
let amplitudeInstance: amplitude.AmplitudeClient | undefined = undefined
type onFinishedType = () => void
const amplitudeLog = (
  name: string,
  eventProperties: any,
  onFinished?: onFinishedType
) => {
  if (!window) return
  if (!(document.domain === "data.ansatt.nav.no")) {
    console.log("Not in prod. Not sending event:", name, eventProperties)
    return
  }
  const AmplitudeConfig = {
    apiEndpoint: 'amplitude.nav.no/collect',
    saveEvents: false,
    includeUtm: true,
    includeReferrer: true,
    platform: window.location.toString(),
  }
  if (!amplitudeInstance) {
    amplitudeInstance = amplitude.getInstance()
    amplitudeInstance.init('2327f913e286c0ca58a41c523b3b5d09', undefined, AmplitudeConfig)
  }
  amplitudeInstance.logEvent(name, eventProperties, onFinished)
}
export default amplitudeLog
