import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { OrganizationService } from '@/services/api'
import type { Organization } from '@/types'

interface OrgContextType {
  currentOrg: Organization | null
  organizations: Organization[]
  setCurrentOrg: (org: Organization | null) => void
  switchOrg: (orgId: string) => void
  loading: boolean
  refreshOrganizations: () => Promise<void>
}

const OrgContext = createContext<OrgContextType | null>(null)

const STORAGE_KEY = 'machineauth_current_org'

export function OrgProvider({ children }: { children: ReactNode }) {
  const [currentOrg, setCurrentOrgState] = useState<Organization | null>(null)
  const [organizations, setOrganizations] = useState<Organization[]>([])
  const [loading, setLoading] = useState(true)

  const refreshOrganizations = async () => {
    try {
      const data = await OrganizationService.list()
      const orgs = data.organizations || []
      setOrganizations(orgs)
      
      // If no current org selected but we have orgs, select first one
      if (!currentOrg && orgs.length > 0) {
        const savedOrgId = localStorage.getItem(STORAGE_KEY)
        const orgToSelect = savedOrgId 
          ? orgs.find((o: Organization) => o.id === savedOrgId) 
          : orgs[0]
        if (orgToSelect) {
          setCurrentOrgState(orgToSelect)
          localStorage.setItem(STORAGE_KEY, orgToSelect.id)
        }
      }
    } catch (error) {
      console.error('Failed to fetch organizations:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refreshOrganizations()
  }, [])

  const setCurrentOrg = (org: Organization | null) => {
    setCurrentOrgState(org)
    if (org) {
      localStorage.setItem(STORAGE_KEY, org.id)
    } else {
      localStorage.removeItem(STORAGE_KEY)
    }
  }

  const switchOrg = (orgId: string) => {
    const org = organizations.find(o => o.id === orgId)
    if (org) {
      setCurrentOrg(org)
    }
  }

  return (
    <OrgContext.Provider value={{ 
      currentOrg, 
      organizations, 
      setCurrentOrg, 
      switchOrg,
      loading,
      refreshOrganizations 
    }}>
      {children}
    </OrgContext.Provider>
  )
}

export function useOrg() {
  const context = useContext(OrgContext)
  if (!context) {
    throw new Error('useOrg must be used within an OrgProvider')
  }
  return context
}
