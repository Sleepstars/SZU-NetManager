import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { ConfigProvider, theme } from 'antd'
import zhCN from 'antd/locale/zh_CN'

type ThemeMode = 'light' | 'dark' | 'auto'

interface ThemeContextType {
  themeMode: ThemeMode
  isDark: boolean
  setThemeMode: (mode: ThemeMode) => void
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined)

export const useTheme = () => {
  const context = useContext(ThemeContext)
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider')
  }
  return context
}

interface ThemeProviderProps {
  children: ReactNode
}

export const ThemeProvider: React.FC<ThemeProviderProps> = ({ children }) => {
  const [themeMode, setThemeMode] = useState<ThemeMode>(() => {
    const saved = localStorage.getItem('theme-mode')
    return (saved as ThemeMode) || 'auto'
  })

  const [systemPrefersDark, setSystemPrefersDark] = useState(() => {
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  })

  const isDark = themeMode === 'dark' || (themeMode === 'auto' && systemPrefersDark)

  useEffect(() => {
    localStorage.setItem('theme-mode', themeMode)
  }, [themeMode])

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = (e: MediaQueryListEvent) => {
      setSystemPrefersDark(e.matches)
    }
    
    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [])

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', isDark ? 'dark' : 'light')
  }, [isDark])

  const antdTheme = {
    algorithm: isDark ? theme.darkAlgorithm : theme.defaultAlgorithm,
    token: {
      colorPrimary: '#722ed1', // 紫色主题
      colorInfo: '#722ed1',
      borderRadius: 8,
      colorBgContainer: isDark ? '#1f1f1f' : '#ffffff',
      colorBorder: isDark ? '#434343' : '#d9d9d9',
      colorText: isDark ? '#ffffffd9' : '#000000d9',
    },
    components: {
      Layout: {
        siderBg: isDark ? '#001529' : '#722ed1',
        headerBg: isDark ? '#001529' : '#ffffff',
      },
      Menu: {
        darkItemBg: '#722ed1',
        darkSubMenuItemBg: '#5a1ea6',
        darkItemSelectedBg: '#9254de',
        darkItemHoverBg: '#9254de',
      },
      Button: {
        colorPrimary: '#722ed1',
        colorPrimaryHover: '#9254de',
        colorPrimaryActive: '#531dab',
      },
      Tag: {
        colorPrimary: '#722ed1',
      }
    },
  }

  return (
    <ThemeContext.Provider value={{ themeMode, isDark, setThemeMode }}>
      <ConfigProvider theme={antdTheme} locale={zhCN}>
        {children}
      </ConfigProvider>
    </ThemeContext.Provider>
  )
}
