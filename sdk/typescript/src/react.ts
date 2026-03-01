/**
 * React hooks for MachineAuth SDK.
 *
 * Usage:
 *   import { MachineAuthProvider, useMachineAuth, useAgents } from '@machineauth/sdk/react';
 *
 * Requires React 18+ as a peer dependency.
 */
import { createContext, useContext, useState, useEffect, useCallback, useMemo } from 'react';
import type { ReactNode } from 'react';
import { MachineAuthClient } from './index';
import type { ClientOptions, Agent, TokenResponse } from './index';

// ── Context ─────────────────────────────────────────────────────

interface MachineAuthContextValue {
  client: MachineAuthClient;
  token: TokenResponse | null;
  authenticate: () => Promise<TokenResponse>;
  logout: () => void;
}

const MachineAuthContext = createContext<MachineAuthContextValue | null>(null);

// ── Provider ────────────────────────────────────────────────────

interface ProviderProps {
  options: ClientOptions;
  children: ReactNode;
}

export function MachineAuthProvider({ options, children }: ProviderProps) {
  const client = useMemo(() => new MachineAuthClient(options), [
    options.baseUrl,
    options.clientId,
    options.clientSecret,
  ]);
  const [token, setToken] = useState<TokenResponse | null>(null);

  const authenticate = useCallback(async () => {
    const tok = await client.clientCredentialsToken();
    setToken(tok);
    return tok;
  }, [client]);

  const logout = useCallback(() => {
    setToken(null);
  }, []);

  const value = useMemo(
    () => ({ client, token, authenticate, logout }),
    [client, token, authenticate, logout],
  );

  return MachineAuthContext.Provider({ value, children } as any);
}

// ── Hooks ───────────────────────────────────────────────────────

export function useMachineAuth(): MachineAuthContextValue {
  const ctx = useContext(MachineAuthContext);
  if (!ctx) {
    throw new Error('useMachineAuth must be used within a MachineAuthProvider');
  }
  return ctx;
}

export function useAgents() {
  const { client } = useMachineAuth();
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const list = await client.listAgents();
      setAgents(list);
    } catch (e) {
      setError(e instanceof Error ? e : new Error(String(e)));
    } finally {
      setLoading(false);
    }
  }, [client]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { agents, loading, error, refresh };
}

export function useToken() {
  const { token, authenticate, logout } = useMachineAuth();
  return { token, authenticate, logout };
}
