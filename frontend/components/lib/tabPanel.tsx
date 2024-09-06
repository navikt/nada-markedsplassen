import * as React from 'react'

export enum TabPanelType {
  vertical = 1,
  simple,
}

interface TabPanelProps {
  children?: React.ReactNode
  index: number
  value: number
  type: TabPanelType
}

const TabPanel = (props: TabPanelProps) => {
  const { children, value, index, ...other } = props
  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`${props.type}-tabpanel-${index}`}
      aria-labelledby={`${props.type}-tab-${index}`}
      style={{ width: '100%' }}
      {...other}
    >
      {value === index && <>{children}</>}
    </div>
  )
}

export default TabPanel
