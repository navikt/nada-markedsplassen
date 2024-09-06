import logo from './Slack-mark-RGB.png'
import Image from 'next/image'

export const SlackLogo = () => (
  <div className="w-8 h-8 flex items-center">
    <Image src={logo} width="24" height="24" alt="Slack"/>
  </div>
)
