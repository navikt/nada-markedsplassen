import type { Preview } from '@storybook/react'
import '../styles/globals.css'
import '@uiw/react-md-editor/markdown-editor.css'
import '@uiw/react-markdown-preview/markdown.css'
import '@navikt/ds-css'
import '@navikt/ds-css-internal'
import '@fontsource/source-sans-pro'
import '@fontsource/source-sans-pro/700.css'

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
}

export default preview
