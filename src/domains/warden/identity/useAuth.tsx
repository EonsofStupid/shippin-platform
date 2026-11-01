// Warden · identity — auth context + hook. Boring, predictable: token in
// localStorage, validated against /me on load.
import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { identityService } from "./identity.svc";
import type { Operator } from "./types";

const TOKEN_KEY = "warden.token";

interface AuthState {
  operator: Operator | null;
  token: string | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [operator, setOperator] = useState<Operator | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem(TOKEN_KEY);
    if (!token) {
      setLoading(false);
      return;
    }
    identityService
      .me(token)
      .then((operator) => {
        setOperator(operator);
        setToken(token);
      })
      .catch(() => localStorage.removeItem(TOKEN_KEY))
      .finally(() => setLoading(false));
  }, []);

  const login = async (username: string, password: string) => {
    const { token, operator } = await identityService.login(username, password);
    localStorage.setItem(TOKEN_KEY, token);
    setOperator(operator);
    setToken(token);
  };

  const logout = async () => {
    const token = localStorage.getItem(TOKEN_KEY);
    localStorage.removeItem(TOKEN_KEY);
    setOperator(null);
    setToken(null);
    if (token) await identityService.logout(token).catch(() => undefined);
  };

  return <AuthContext.Provider value={{ operator, token, loading, login, logout }}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
