"""Async MachineAuth SDK client using httpx."""
from __future__ import annotations

import json
import os
from typing import Any, Dict, Iterable, List, Optional

try:
    import httpx
except ImportError:
    httpx = None  # type: ignore[assignment]

from .models import (
    Agent,
    AgentUsage,
    APIKey,
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
    ) -> None:
        super().__init__(message)
        self.status_code = status_code
        self.body = body


class AsyncMachineAuthClient:
    """Async MachineAuth SDK client with the same API surface as the sync client."""

    def __init__(
        self,
        base_url: str = "http://localhost:8081",
        client_id: Optional[str] = None,
        client_secret: Optional[str] = None,
        access_token: Optional[str] = None,
        timeout: float = 10.0,
        http_client: Optional[Any] = None,
    ) -> None:
        if httpx is None:
            raise ImportError("httpx is required for AsyncMachineAuthClient: pip install httpx")
        self.base_url = base_url.rstrip("/")
        self.client_id = client_id
        self.client_secret = client_secret
        self.access_token = access_token
        self.timeout = timeout
        self._client = http_client or httpx.AsyncClient(timeout=timeout)
        self._owns_client = http_client is None

    @classmethod
    def from_env(cls) -> "AsyncMachineAuthClient":
        return cls(
            base_url=os.environ.get("MACHINEAUTH_BASE_URL", "http://localhost:8081"),
            client_id=os.environ.get("MACHINEAUTH_CLIENT_ID"),
            client_secret=os.environ.get("MACHINEAUTH_CLIENT_SECRET"),
            access_token=os.environ.get("MACHINEAUTH_ACCESS_TOKEN"),
        )

    async def __aenter__(self) -> "AsyncMachineAuthClient":
        return self

    async def __aexit__(self, *args: Any) -> None:
        await self.close()

    async def close(self) -> None:
        if self._owns_client:
            await self._client.aclose()

    def with_token(self, token: str) -> "AsyncMachineAuthClient":
        self.access_token = token
        return self

    # ── Token operations ─────────────────────────────────────────

    async def client_credentials_token(
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
        raw = await self._request("POST", "/oauth/token", data=data)
        token = TokenResponse.from_dict(raw)
        if token.access_token:
            self.access_token = token.access_token
        return token

    async def refresh_token(
        self,
        refresh_token: str,
        client_id: Optional[str] = None,
        client_secret: Optional[str] = None,
    ) -> TokenResponse:
        data: Dict[str, str] = {"grant_type": "refresh_token", "refresh_token": refresh_token}
        if client_id or self.client_id:
            data["client_id"] = client_id or self.client_id  # type: ignore[assignment]
        if client_secret or self.client_secret:
            data["client_secret"] = client_secret or self.client_secret  # type: ignore[assignment]
        raw = await self._request("POST", "/oauth/token", data=data)
        token = TokenResponse.from_dict(raw)
        if token.access_token:
            self.access_token = token.access_token
        return token

    async def introspect(self, token: str) -> Dict[str, Any]:
        return await self._request("POST", "/oauth/introspect", data={"token": token})

    async def revoke(self, token: str, token_type_hint: Optional[str] = None) -> None:
        data: Dict[str, str] = {"token": token}
        if token_type_hint:
            data["token_type_hint"] = token_type_hint
        await self._request("POST", "/oauth/revoke", data=data)

    # ── Agent management ─────────────────────────────────────────

    async def list_agents(self, token: Optional[str] = None) -> List[Agent]:
        raw = await self._request("GET", "/api/agents", token=token)
        agents = raw.get("agents", raw) if isinstance(raw, dict) else raw
        return [Agent.from_dict(a) for a in agents] if isinstance(agents, list) else []

    async def create_agent(
        self, name: str, scopes: Optional[Iterable[str]] = None,
        organization_id: Optional[str] = None, team_id: Optional[str] = None,
        expires_in: Optional[int] = None, token: Optional[str] = None,
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
        return await self._request("POST", "/api/agents", json_body=payload, token=token)

    async def get_agent(self, agent_id: str, token: Optional[str] = None) -> Agent:
        raw = await self._request("GET", f"/api/agents/{agent_id}", token=token)
        return Agent.from_dict(raw.get("agent", raw) if isinstance(raw, dict) else raw)

    async def delete_agent(self, agent_id: str, token: Optional[str] = None) -> None:
        await self._request("DELETE", f"/api/agents/{agent_id}", token=token)

    async def rotate_agent(self, agent_id: str, token: Optional[str] = None) -> Dict[str, Any]:
        return await self._request("POST", f"/api/agents/{agent_id}/rotate", token=token)

    # ── Self-service ─────────────────────────────────────────────

    async def get_me(self, token: Optional[str] = None) -> Agent:
        raw = await self._request("GET", "/api/agents/me", token=token)
        return Agent.from_dict(raw.get("agent", raw) if isinstance(raw, dict) else raw)

    async def get_my_usage(self, token: Optional[str] = None) -> AgentUsage:
        raw = await self._request("GET", "/api/agents/me/usage", token=token)
        return AgentUsage.from_dict(raw)

    async def rotate_me(self, token: Optional[str] = None) -> Dict[str, Any]:
        return await self._request("POST", "/api/agents/me/rotate", token=token)

    async def deactivate_me(self, token: Optional[str] = None) -> Dict[str, Any]:
        return await self._request("POST", "/api/agents/me/deactivate", token=token)

    async def reactivate_me(self, token: Optional[str] = None) -> Dict[str, Any]:
        return await self._request("POST", "/api/agents/me/reactivate", token=token)

    async def delete_me(self, token: Optional[str] = None) -> None:
        await self._request("DELETE", "/api/agents/me", token=token)

    # ── Organizations ────────────────────────────────────────────

    async def list_organizations(self, token: Optional[str] = None) -> List[Organization]:
        raw = await self._request("GET", "/api/organizations", token=token)
        orgs = raw.get("organizations", raw) if isinstance(raw, dict) else raw
        return [Organization.from_dict(o) for o in orgs] if isinstance(orgs, list) else []

    async def create_organization(self, name: str, slug: str, owner_email: str, token: Optional[str] = None) -> Organization:
        return Organization.from_dict(
            await self._request("POST", "/api/organizations", json_body={"name": name, "slug": slug, "owner_email": owner_email}, token=token)
        )

    async def get_organization(self, org_id: str, token: Optional[str] = None) -> Organization:
        return Organization.from_dict(await self._request("GET", f"/api/organizations/{org_id}", token=token))

    async def update_organization(self, org_id: str, token: Optional[str] = None, **kwargs: Any) -> Organization:
        return Organization.from_dict(await self._request("PUT", f"/api/organizations/{org_id}", json_body=kwargs, token=token))

    async def delete_organization(self, org_id: str, token: Optional[str] = None) -> None:
        await self._request("DELETE", f"/api/organizations/{org_id}", token=token)

    # ── Teams ────────────────────────────────────────────────────

    async def list_teams(self, org_id: str, token: Optional[str] = None) -> List[Team]:
        raw = await self._request("GET", f"/api/organizations/{org_id}/teams", token=token)
        teams = raw.get("teams", raw) if isinstance(raw, dict) else raw
        return [Team.from_dict(t) for t in teams] if isinstance(teams, list) else []

    async def create_team(self, org_id: str, name: str, description: Optional[str] = None, token: Optional[str] = None) -> Team:
        payload: Dict[str, Any] = {"name": name}
        if description:
            payload["description"] = description
        return Team.from_dict(await self._request("POST", f"/api/organizations/{org_id}/teams", json_body=payload, token=token))

    # ── API Keys ─────────────────────────────────────────────────

    async def list_api_keys(self, org_id: str, token: Optional[str] = None) -> List[APIKey]:
        raw = await self._request("GET", f"/api/organizations/{org_id}/api-keys", token=token)
        keys = raw.get("api_keys", raw) if isinstance(raw, dict) else raw
        return [APIKey.from_dict(k) for k in keys] if isinstance(keys, list) else []

    async def create_api_key(self, org_id: str, name: str, team_id: Optional[str] = None, expires_in: Optional[int] = None, token: Optional[str] = None) -> Dict[str, Any]:
        payload: Dict[str, Any] = {"name": name}
        if team_id:
            payload["team_id"] = team_id
        if expires_in is not None:
            payload["expires_in"] = expires_in
        return await self._request("POST", f"/api/organizations/{org_id}/api-keys", json_body=payload, token=token)

    async def delete_api_key(self, org_id: str, key_id: str, token: Optional[str] = None) -> None:
        await self._request("DELETE", f"/api/organizations/{org_id}/api-keys/{key_id}", token=token)

    # ── Webhooks ─────────────────────────────────────────────────

    async def list_webhooks(self, token: Optional[str] = None) -> List[WebhookConfig]:
        raw = await self._request("GET", "/api/webhooks", token=token)
        hooks = raw.get("webhooks", raw) if isinstance(raw, dict) else raw
        return [WebhookConfig.from_dict(w) for w in hooks] if isinstance(hooks, list) else []

    async def create_webhook(self, name: str, url: str, events: List[str], max_retries: Optional[int] = None, retry_backoff_base: Optional[int] = None, token: Optional[str] = None) -> Dict[str, Any]:
        payload: Dict[str, Any] = {"name": name, "url": url, "events": events}
        if max_retries is not None:
            payload["max_retries"] = max_retries
        if retry_backoff_base is not None:
            payload["retry_backoff_base"] = retry_backoff_base
        return await self._request("POST", "/api/webhooks", json_body=payload, token=token)

    async def get_webhook(self, webhook_id: str, token: Optional[str] = None) -> WebhookConfig:
        raw = await self._request("GET", f"/api/webhooks/{webhook_id}", token=token)
        return WebhookConfig.from_dict(raw.get("webhook", raw) if isinstance(raw, dict) else raw)

    async def update_webhook(self, webhook_id: str, token: Optional[str] = None, **kwargs: Any) -> WebhookConfig:
        raw = await self._request("PUT", f"/api/webhooks/{webhook_id}", json_body=kwargs, token=token)
        return WebhookConfig.from_dict(raw.get("webhook", raw) if isinstance(raw, dict) else raw)

    async def delete_webhook(self, webhook_id: str, token: Optional[str] = None) -> None:
        await self._request("DELETE", f"/api/webhooks/{webhook_id}", token=token)

    async def test_webhook(self, webhook_id: str, event: str, payload: Optional[Dict[str, Any]] = None, token: Optional[str] = None) -> Dict[str, Any]:
        body: Dict[str, Any] = {"event": event}
        if payload:
            body["payload"] = payload
        return await self._request("POST", f"/api/webhooks/{webhook_id}/test", json_body=body, token=token)

    async def list_deliveries(self, webhook_id: str, token: Optional[str] = None) -> List[WebhookDelivery]:
        raw = await self._request("GET", f"/api/webhooks/{webhook_id}/deliveries", token=token)
        deliveries = raw.get("deliveries", raw) if isinstance(raw, dict) else raw
        return [WebhookDelivery.from_dict(d) for d in deliveries] if isinstance(deliveries, list) else []

    async def get_delivery(self, webhook_id: str, delivery_id: str, token: Optional[str] = None) -> WebhookDelivery:
        raw = await self._request("GET", f"/api/webhooks/{webhook_id}/deliveries/{delivery_id}", token=token)
        return WebhookDelivery.from_dict(raw.get("delivery", raw) if isinstance(raw, dict) else raw)

    async def list_webhook_events(self, token: Optional[str] = None) -> List[str]:
        raw = await self._request("GET", "/api/webhook-events", token=token)
        return raw.get("events", []) if isinstance(raw, dict) else []

    # ── Internal ─────────────────────────────────────────────────

    def _auth_headers(self, token: Optional[str]) -> Dict[str, str]:
        bearer = token or self.access_token
        return {"Authorization": f"Bearer {bearer}"} if bearer else {}

    async def _request(
        self,
        method: str,
        path: str,
        *,
        data: Optional[Dict[str, str]] = None,
        json_body: Optional[Dict[str, Any]] = None,
        token: Optional[str] = None,
    ) -> Any:
        url = f"{self.base_url}{path}"
        headers: Dict[str, str] = {"Accept": "application/json", **self._auth_headers(token)}
        kwargs: Dict[str, Any] = {"headers": headers}

        if data is not None:
            headers["Content-Type"] = "application/x-www-form-urlencoded"
            kwargs["data"] = data
        elif json_body is not None:
            headers["Content-Type"] = "application/json"
            kwargs["content"] = json.dumps(json_body).encode()

        response = await self._client.request(method, url, **kwargs)
        content_type = response.headers.get("content-type", "")

        body: Any = None
        if "application/json" in content_type:
            try:
                body = response.json()
            except (json.JSONDecodeError, ValueError):
                body = response.text
        else:
            body = response.text

        if response.status_code >= 400:
            message = body.get("message") if isinstance(body, dict) else str(response.status_code)
            raise MachineAuthError(response.status_code, message, body)
        return body
