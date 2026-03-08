from __future__ import annotations

import json
import os
import threading
import time
from typing import Any, Dict, Iterable, List, Optional

import requests

from .models import (
    Agent,
    AgentUsage,
    APIKey,
    AutoRefreshOptions,
    Organization,
    Team,
    TokenResponse,
    WebhookConfig,
    WebhookDelivery,
)


class MachineAuthError(Exception):
    """Represents an error response from the MachineAuth API."""

    def __init__(
        self,
        status_code: int,
        message: str,
        body: Any | None = None,
        response: Optional[requests.Response] = None,
    ) -> None:
        super().__init__(message)
        self.status_code = status_code
        self.body = body
        self.response = response


class MachineAuthClient:
    """MachineAuth SDK client covering the full API surface."""

    def __init__(
        self,
        base_url: str = "http://localhost:8081",
        client_id: Optional[str] = None,
        client_secret: Optional[str] = None,
        access_token: Optional[str] = None,
        timeout: float = 10.0,
        session: Optional[requests.Session] = None,
        auto_refresh: Optional[AutoRefreshOptions] = None,
    ) -> None:
        self.base_url = base_url.rstrip("/")
        self.client_id = client_id
        self.client_secret = client_secret
        self.access_token = access_token
        self.timeout = timeout
        self.session = session or requests.Session()
        self._owns_session = session is None
        self._auto_refresh = auto_refresh
        self._cached_token: Optional[TokenResponse] = None
        self._token_expiry_time: Optional[float] = None
        self._refresh_timer: Optional[threading.Timer] = None
        self._refresh_lock = threading.Lock()

    # ── Auto-refresh helpers ─────────────────────────────────────────

    def _is_token_expiring_soon(self) -> bool:
        if not self._auto_refresh or not self._auto_refresh.enabled:
            return False
        if not self._token_expiry_time:
            return False
        buffer = self._auto_refresh.refresh_buffer_seconds
        return (time.time() + buffer) >= self._token_expiry_time

    def _set_cached_token(self, token: TokenResponse) -> None:
        self._cached_token = token
        self.access_token = token.access_token
        expires_in = token.expires_in or 3600
        self._token_expiry_time = time.time() + expires_in
        self._schedule_auto_refresh(expires_in)

    def _schedule_auto_refresh(self, expires_in: int) -> None:
        if self._refresh_timer:
            self._refresh_timer.cancel()
        if not self._auto_refresh or not self._auto_refresh.enabled:
            return
        buffer = self._auto_refresh.refresh_buffer_seconds
        refresh_delay = expires_in - buffer
        if refresh_delay > 0:
            self._refresh_timer = threading.Timer(refresh_delay, self._do_refresh_token)
            self._refresh_timer.daemon = True
            self._refresh_timer.start()

    def _do_refresh_token(self) -> TokenResponse:
        if not self._auto_refresh or not self._auto_refresh.enabled:
            raise ValueError("Auto-refresh is not enabled")
        if not self.client_id or not self.client_secret:
            raise ValueError("client_id and client_secret are required for auto-refresh")

        data: Dict[str, str] = {
            "grant_type": "client_credentials",
            "client_id": self.client_id,
            "client_secret": self.client_secret,
        }
        if self._cached_token and self._cached_token.scope:
            data["scope"] = self._cached_token.scope

        raw = self._request(
            "POST",
            "/oauth/token",
            data=data,
            headers={"Content-Type": "application/x-www-form-urlencoded"},
        )
        token = TokenResponse.from_dict(raw)
        self._set_cached_token(token)
        return token

    def _refresh_token_if_needed(self, token: Optional[str]) -> Optional[str]:
        if token or not self._auto_refresh or not self._auto_refresh.enabled or not self._cached_token:
            return token

        if not self._is_token_expiring_soon():
            return self._cached_token.access_token

        with self._refresh_lock:
            if not self._is_token_expiring_soon():
                return self._cached_token.access_token
            return self._do_refresh_token().access_token

    def force_token_refresh(self) -> TokenResponse:
        """Force an immediate token refresh."""
        return self._do_refresh_token()

    def get_cached_token(self) -> Optional[TokenResponse]:
        """Get the current cached token if available."""
        return self._cached_token

    def is_auto_refresh_enabled(self) -> bool:
        """Check if auto-refresh is enabled."""
        return self._auto_refresh.enabled if self._auto_refresh else False

    @classmethod
    def from_env(cls) -> "MachineAuthClient":
        """Create a client from MACHINEAUTH_* environment variables."""
        return cls(
            base_url=os.environ.get("MACHINEAUTH_BASE_URL", "http://localhost:8081"),
            client_id=os.environ.get("MACHINEAUTH_CLIENT_ID"),
            client_secret=os.environ.get("MACHINEAUTH_CLIENT_SECRET"),
            access_token=os.environ.get("MACHINEAUTH_ACCESS_TOKEN"),
        )

    def __enter__(self) -> "MachineAuthClient":
        return self

    def __exit__(self, *args: Any) -> None:
        self.close()

    def close(self) -> None:
        """Close the underlying session if we own it."""
        if self._owns_session:
            self.session.close()

    def with_token(self, token: str) -> "MachineAuthClient":
        self.access_token = token
        return self

    # ── Token operations ─────────────────────────────────────────────

    def client_credentials_token(
        self,
        client_id: Optional[str] = None,
        client_secret: Optional[str] = None,
        scope: Optional[Iterable[str]] = None,
    ) -> TokenResponse:
        cid = client_id or self.client_id
        csec = client_secret or self.client_secret
        if not cid or not csec:
            raise ValueError("client_id and client_secret are required")

        data: Dict[str, str] = {
            "grant_type": "client_credentials",
            "client_id": cid,
            "client_secret": csec,
        }
        if scope:
            data["scope"] = " ".join(scope)

        raw = self._request(
            "POST",
            "/oauth/token",
            data=data,
            headers={"Content-Type": "application/x-www-form-urlencoded"},
        )
        token = TokenResponse.from_dict(raw)
        self._set_cached_token(token)
        return token

    def refresh_token(
        self,
        refresh_token: str,
        client_id: Optional[str] = None,
        client_secret: Optional[str] = None,
    ) -> TokenResponse:
        data: Dict[str, str] = {
            "grant_type": "refresh_token",
            "refresh_token": refresh_token,
        }
        if client_id or self.client_id:
            data["client_id"] = client_id or self.client_id  # type: ignore[assignment]
        if client_secret or self.client_secret:
            data["client_secret"] = client_secret or self.client_secret  # type: ignore[assignment]

        raw = self._request(
            "POST",
            "/oauth/token",
            data=data,
            headers={"Content-Type": "application/x-www-form-urlencoded"},
        )
        token = TokenResponse.from_dict(raw)
        if token.access_token:
            self.access_token = token.access_token
        return token

    def introspect(self, token: str) -> Dict[str, Any]:
        return self._request(
            "POST",
            "/oauth/introspect",
            data={"token": token},
            headers={"Content-Type": "application/x-www-form-urlencoded"},
        )

    def revoke(self, token: str, token_type_hint: Optional[str] = None) -> None:
        data: Dict[str, str] = {"token": token}
        if token_type_hint:
            data["token_type_hint"] = token_type_hint
        self._request(
            "POST",
            "/oauth/revoke",
            data=data,
            headers={"Content-Type": "application/x-www-form-urlencoded"},
        )

    # ── Agent management ─────────────────────────────────────────────

    def list_agents(self, token: Optional[str] = None) -> List[Agent]:
        raw = self._request("GET", "/api/agents", headers=self._auth_headers(token))
        agents = raw.get("agents", raw) if isinstance(raw, dict) else raw
        return [Agent.from_dict(a) for a in agents] if isinstance(agents, list) else []

    def create_agent(
        self,
        name: str,
        scopes: Optional[Iterable[str]] = None,
        organization_id: Optional[str] = None,
        team_id: Optional[str] = None,
        expires_in: Optional[int] = None,
        token: Optional[str] = None,
    ) -> Dict[str, Any]:
        payload: Dict[str, Any] = {"name": name}
        if scopes:
            payload["scopes"] = list(scopes)
        if organization_id:
            payload["organization_id"] = organization_id
        if team_id:
            payload["team_id"] = team_id
        if expires_in is not None:
            payload["expires_in"] = expires_in

        return self._request(
            "POST",
            "/api/agents",
            json=payload,
            headers={**self._auth_headers(token), "Content-Type": "application/json"},
        )

    def get_agent(self, agent_id: str, token: Optional[str] = None) -> Agent:
        raw = self._request(
            "GET",
            f"/api/agents/{agent_id}",
            headers=self._auth_headers(token),
        )
        agent_data = raw.get("agent", raw) if isinstance(raw, dict) else raw
        return Agent.from_dict(agent_data)

    def delete_agent(self, agent_id: str, token: Optional[str] = None) -> None:
        self._request(
            "DELETE",
            f"/api/agents/{agent_id}",
            headers=self._auth_headers(token),
        )

    def rotate_agent(self, agent_id: str, token: Optional[str] = None) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/api/agents/{agent_id}/rotate",
            headers=self._auth_headers(token),
        )

    # ── Self-service (agent-as-caller) ───────────────────────────────

    def get_me(self, token: Optional[str] = None) -> Agent:
        raw = self._request("GET", "/api/agents/me", headers=self._auth_headers(token))
        agent_data = raw.get("agent", raw) if isinstance(raw, dict) else raw
        return Agent.from_dict(agent_data)

    def get_my_usage(self, token: Optional[str] = None) -> AgentUsage:
        raw = self._request("GET", "/api/agents/me/usage", headers=self._auth_headers(token))
        return AgentUsage.from_dict(raw)

    def rotate_me(self, token: Optional[str] = None) -> Dict[str, Any]:
        return self._request("POST", "/api/agents/me/rotate", headers=self._auth_headers(token))

    def deactivate_me(self, token: Optional[str] = None) -> Dict[str, Any]:
        return self._request("POST", "/api/agents/me/deactivate", headers=self._auth_headers(token))

    def reactivate_me(self, token: Optional[str] = None) -> Dict[str, Any]:
        return self._request("POST", "/api/agents/me/reactivate", headers=self._auth_headers(token))

    def delete_me(self, token: Optional[str] = None) -> None:
        self._request("DELETE", "/api/agents/me", headers=self._auth_headers(token))

    # ── Organizations ────────────────────────────────────────────────

    def list_organizations(self, token: Optional[str] = None) -> List[Organization]:
        raw = self._request("GET", "/api/organizations", headers=self._auth_headers(token))
        orgs = raw.get("organizations", raw) if isinstance(raw, dict) else raw
        return [Organization.from_dict(o) for o in orgs] if isinstance(orgs, list) else []

    def create_organization(
        self,
        name: str,
        slug: str,
        owner_email: str,
        token: Optional[str] = None,
    ) -> Organization:
        raw = self._request(
            "POST",
            "/api/organizations",
            json={"name": name, "slug": slug, "owner_email": owner_email},
            headers={**self._auth_headers(token), "Content-Type": "application/json"},
        )
        return Organization.from_dict(raw)

    def get_organization(self, org_id: str, token: Optional[str] = None) -> Organization:
        raw = self._request(
            "GET", f"/api/organizations/{org_id}", headers=self._auth_headers(token)
        )
        return Organization.from_dict(raw)

    def update_organization(
        self,
        org_id: str,
        token: Optional[str] = None,
        **kwargs: Any,
    ) -> Organization:
        raw = self._request(
            "PUT",
            f"/api/organizations/{org_id}",
            json=kwargs,
            headers={**self._auth_headers(token), "Content-Type": "application/json"},
        )
        return Organization.from_dict(raw)

    def delete_organization(self, org_id: str, token: Optional[str] = None) -> None:
        self._request(
            "DELETE", f"/api/organizations/{org_id}", headers=self._auth_headers(token)
        )

    # ── Teams ────────────────────────────────────────────────────────

    def list_teams(self, org_id: str, token: Optional[str] = None) -> List[Team]:
        raw = self._request(
            "GET", f"/api/organizations/{org_id}/teams", headers=self._auth_headers(token)
        )
        teams = raw.get("teams", raw) if isinstance(raw, dict) else raw
        return [Team.from_dict(t) for t in teams] if isinstance(teams, list) else []

    def create_team(
        self,
        org_id: str,
        name: str,
        description: Optional[str] = None,
        token: Optional[str] = None,
    ) -> Team:
        payload: Dict[str, Any] = {"name": name}
        if description:
            payload["description"] = description
        raw = self._request(
            "POST",
            f"/api/organizations/{org_id}/teams",
            json=payload,
            headers={**self._auth_headers(token), "Content-Type": "application/json"},
        )
        return Team.from_dict(raw)

    # ── API Keys ─────────────────────────────────────────────────────

    def list_api_keys(self, org_id: str, token: Optional[str] = None) -> List[APIKey]:
        raw = self._request(
            "GET", f"/api/organizations/{org_id}/api-keys", headers=self._auth_headers(token)
        )
        keys = raw.get("api_keys", raw) if isinstance(raw, dict) else raw
        return [APIKey.from_dict(k) for k in keys] if isinstance(keys, list) else []

    def create_api_key(
        self,
        org_id: str,
        name: str,
        team_id: Optional[str] = None,
        expires_in: Optional[int] = None,
        token: Optional[str] = None,
    ) -> Dict[str, Any]:
        payload: Dict[str, Any] = {"name": name}
        if team_id:
            payload["team_id"] = team_id
        if expires_in is not None:
            payload["expires_in"] = expires_in
        return self._request(
            "POST",
            f"/api/organizations/{org_id}/api-keys",
            json=payload,
            headers={**self._auth_headers(token), "Content-Type": "application/json"},
        )

    def delete_api_key(
        self, org_id: str, key_id: str, token: Optional[str] = None
    ) -> None:
        self._request(
            "DELETE",
            f"/api/organizations/{org_id}/api-keys/{key_id}",
            headers=self._auth_headers(token),
        )

    # ── Webhooks ─────────────────────────────────────────────────────

    def list_webhooks(self, token: Optional[str] = None) -> List[WebhookConfig]:
        raw = self._request("GET", "/api/webhooks", headers=self._auth_headers(token))
        hooks = raw.get("webhooks", raw) if isinstance(raw, dict) else raw
        return [WebhookConfig.from_dict(w) for w in hooks] if isinstance(hooks, list) else []

    def create_webhook(
        self,
        name: str,
        url: str,
        events: List[str],
        max_retries: Optional[int] = None,
        retry_backoff_base: Optional[int] = None,
        token: Optional[str] = None,
    ) -> Dict[str, Any]:
        payload: Dict[str, Any] = {"name": name, "url": url, "events": events}
        if max_retries is not None:
            payload["max_retries"] = max_retries
        if retry_backoff_base is not None:
            payload["retry_backoff_base"] = retry_backoff_base
        return self._request(
            "POST",
            "/api/webhooks",
            json=payload,
            headers={**self._auth_headers(token), "Content-Type": "application/json"},
        )

    def get_webhook(self, webhook_id: str, token: Optional[str] = None) -> WebhookConfig:
        raw = self._request(
            "GET", f"/api/webhooks/{webhook_id}", headers=self._auth_headers(token)
        )
        hook_data = raw.get("webhook", raw) if isinstance(raw, dict) else raw
        return WebhookConfig.from_dict(hook_data)

    def update_webhook(
        self,
        webhook_id: str,
        token: Optional[str] = None,
        **kwargs: Any,
    ) -> WebhookConfig:
        raw = self._request(
            "PUT",
            f"/api/webhooks/{webhook_id}",
            json=kwargs,
            headers={**self._auth_headers(token), "Content-Type": "application/json"},
        )
        hook_data = raw.get("webhook", raw) if isinstance(raw, dict) else raw
        return WebhookConfig.from_dict(hook_data)

    def delete_webhook(self, webhook_id: str, token: Optional[str] = None) -> None:
        self._request(
            "DELETE", f"/api/webhooks/{webhook_id}", headers=self._auth_headers(token)
        )

    def test_webhook(
        self,
        webhook_id: str,
        event: str,
        payload: Optional[Dict[str, Any]] = None,
        token: Optional[str] = None,
    ) -> Dict[str, Any]:
        body: Dict[str, Any] = {"event": event}
        if payload:
            body["payload"] = payload
        return self._request(
            "POST",
            f"/api/webhooks/{webhook_id}/test",
            json=body,
            headers={**self._auth_headers(token), "Content-Type": "application/json"},
        )

    def list_deliveries(
        self, webhook_id: str, token: Optional[str] = None
    ) -> List[WebhookDelivery]:
        raw = self._request(
            "GET",
            f"/api/webhooks/{webhook_id}/deliveries",
            headers=self._auth_headers(token),
        )
        deliveries = raw.get("deliveries", raw) if isinstance(raw, dict) else raw
        return (
            [WebhookDelivery.from_dict(d) for d in deliveries]
            if isinstance(deliveries, list)
            else []
        )

    def get_delivery(
        self, webhook_id: str, delivery_id: str, token: Optional[str] = None
    ) -> WebhookDelivery:
        raw = self._request(
            "GET",
            f"/api/webhooks/{webhook_id}/deliveries/{delivery_id}",
            headers=self._auth_headers(token),
        )
        data = raw.get("delivery", raw) if isinstance(raw, dict) else raw
        return WebhookDelivery.from_dict(data)

    def list_webhook_events(self, token: Optional[str] = None) -> List[str]:
        raw = self._request(
            "GET", "/api/webhook-events", headers=self._auth_headers(token)
        )
        return raw.get("events", []) if isinstance(raw, dict) else []

    # ── Internal helpers ─────────────────────────────────────────────

    def _auth_headers(self, token: Optional[str]) -> Dict[str, str]:
        bearer = token or self.access_token
        return {"Authorization": f"Bearer {bearer}"} if bearer else {}

    def _request(
        self,
        method: str,
        path: str,
        *,
        headers: Optional[Dict[str, str]] = None,
        **kwargs: Any,
    ) -> Any:
        # Auto-refresh token if needed before making the request
        if self._auto_refresh and self._auto_refresh.enabled:
            has_auth = headers and "Authorization" in headers
            if not has_auth:
                token = self._refresh_token_if_needed(None)
                if token:
                    headers = {**(headers or {}), f"Authorization": f"Bearer {token}"}

        url = f"{self.base_url}{path}"
        merged_headers = {"Accept": "application/json", **(headers or {})}
        response = self.session.request(
            method, url, headers=merged_headers, timeout=self.timeout, **kwargs
        )
        content_type = response.headers.get("content-type", "")

        body: Any = None
        if "application/json" in content_type:
            try:
                body = response.json()
            except json.JSONDecodeError:
                body = response.text
        else:
            body = response.text

        if not response.ok:
            message = body.get("message") if isinstance(body, dict) else response.reason
            raise MachineAuthError(response.status_code, message, body, response)
        return body
