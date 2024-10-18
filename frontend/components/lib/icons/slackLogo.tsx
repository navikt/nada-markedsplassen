import logo from './Slack-mark-RGB.png'
import Image from 'next/image'

export const SlackLogo = () => (
  <div className="w-8 h-8 flex items-center">
    <Image src={logo} width="24" height="24" alt="Slack"/>
  </div>
)

export const SlackLogoMini = () => (
  <div className="w-4 h-4 flex pl-0 pr-0 items-center">
    <Image src={logo} width="14" height="14" alt="Slack"/>
  </div>
)