"""Unit tests for MachineAuth Python SDK."""
from __future__ import annotations

import json
from typing import Any, Dict
from unittest.mock import MagicMock

import pytest

from machineauth import MachineAuthClient, MachineAuthError
from machineauth.models import (
    Agent,
    AgentUsage,
    APIKey,
    Organization,
    Team,
    TokenResponse,
    WebhookConfig,
    WebhookDelivery,
)


def _mock_response(
    status_code: int = 200,
    json_data: Any = None,
    text: str = "",
    content_type: str = "application/json",
    reason: str = "OK",
) -> MagicMock:
    resp = MagicMock()
    resp.status_code = status_code
    resp.ok = 200 <= status_code < 300
    resp.reason = reason
    resp.headers = {"content-type": content_type}
    resp.text = text or (json.dumps(json_data) if json_data is not None else "")
    resp.json.return_value = json_data
    return resp


# ── Token operations ─────────────────────────────────────────────────


class TestTokenOperations:
    def test_client_credentials_token(self) -> None:
        token_data = {"access_token": "tok-123", "token_type": "Bearer", "expires_in": 3600}
        session = MagicMock()
        session.request.return_value = _mock_response(json_data=token_data)

        client = MachineAuthClient(client_id="cid", client_secret="csec", session=session)
        result = client.client_credentials_token()

        assert isinstance(result, TokenResponse)
        assert result.access_token == "tok-123"
        assert client.access_token == "tok-123"

    def test_refresh_token(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"access_token": "tok-new", "token_type": "Bearer"}
        )
        client = MachineAuthClient(session=session)
        result = client.refresh_token("refresh-abc")
        assert result.access_token == "tok-new"

    def test_introspect(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(json_data={"active": True, "scope": "read"})
        client = MachineAuthClient(session=session)
        assert client.introspect("tok-123")["active"] is True

    def test_revoke(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(json_data={"status": "revoked"})
        client = MachineAuthClient(session=session)
        client.revoke("tok-123")
        session.request.assert_called_once()

    def test_missing_credentials_raises(self) -> None:
        with pytest.raises(ValueError, match="client_id and client_secret"):
            MachineAuthClient().client_credentials_token()


# ── Agent management ─────────────────────────────────────────────────


def _agent_dict(**overrides: Any) -> Dict[str, Any]:
    base: Dict[str, Any] = {"id": "a-1", "name": "test-agent", "client_id": "cid-1", "scopes": ["read"]}
    base.update(overrides)
    return base


class TestAgentManagement:
    def test_list_agents(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(json_data={"agents": [_agent_dict()]})
        agents = MachineAuthClient(session=session).list_agents()
        assert len(agents) == 1 and isinstance(agents[0], Agent)

    def test_create_agent(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"agent": _agent_dict(), "client_secret": "sec"}
        )
        result = MachineAuthClient(session=session).create_agent("test-agent", scopes=["read"])
        assert result["client_secret"] == "sec"

    def test_get_agent(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(json_data={"agent": _agent_dict()})
        agent = MachineAuthClient(session=session).get_agent("a-1")
        assert isinstance(agent, Agent) and agent.id == "a-1"

    def test_delete_agent(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(status_code=204, text="")
        MachineAuthClient(session=session).delete_agent("a-1")
        session.request.assert_called_once()

    def test_rotate_agent(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(json_data={"client_secret": "new-sec"})
        assert MachineAuthClient(session=session).rotate_agent("a-1")["client_secret"] == "new-sec"


# ── Self-service ─────────────────────────────────────────────────────


class TestSelfService:
    def test_get_me(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"agent": {"id": "me", "name": "self", "client_id": "c"}}
        )
        agent = MachineAuthClient(access_token="tok", session=session).get_me()
        assert isinstance(agent, Agent) and agent.id == "me"

    def test_get_my_usage(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"agent": {"id": "me", "name": "self", "client_id": "c"}, "token_count": 42}
        )
        usage = MachineAuthClient(access_token="tok", session=session).get_my_usage()
        assert isinstance(usage, AgentUsage) and usage.token_count == 42

    def test_rotate_me(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(json_data={"client_secret": "rotated"})
        assert MachineAuthClient(access_token="tok", session=session).rotate_me()["client_secret"] == "rotated"

    def test_deactivate_reactivate_me(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(json_data={"message": "ok"})
        client = MachineAuthClient(access_token="tok", session=session)
        assert "message" in client.deactivate_me()
        assert "message" in client.reactivate_me()

    def test_delete_me(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(status_code=204, text="")
        MachineAuthClient(access_token="tok", session=session).delete_me()
        session.request.assert_called_once()


# ── Organizations ────────────────────────────────────────────────────


class TestOrganizations:
    def test_list_organizations(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"organizations": [{"id": "o1", "name": "Org", "slug": "org"}]}
        )
        orgs = MachineAuthClient(session=session).list_organizations()
        assert len(orgs) == 1 and isinstance(orgs[0], Organization)

    def test_create_organization(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"id": "o1", "name": "Org", "slug": "org"}
        )
        org = MachineAuthClient(session=session).create_organization("Org", "org", "a@b.com")
        assert isinstance(org, Organization) and org.id == "o1"

    def test_update_organization(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"id": "o1", "name": "Updated", "slug": "org"}
        )
        org = MachineAuthClient(session=session).update_organization("o1", name="Updated")
        assert org.name == "Updated"

    def test_delete_organization(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(status_code=204, text="")
        MachineAuthClient(session=session).delete_organization("o1")
        session.request.assert_called_once()


# ── Teams ────────────────────────────────────────────────────────────


class TestTeams:
    def test_list_teams(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"teams": [{"id": "t1", "name": "T", "organization_id": "o1"}]}
        )
        teams = MachineAuthClient(session=session).list_teams("o1")
        assert len(teams) == 1 and isinstance(teams[0], Team)

    def test_create_team(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"id": "t1", "name": "T", "organization_id": "o1"}
        )
        team = MachineAuthClient(session=session).create_team("o1", "T")
        assert isinstance(team, Team)


# ── API Keys ─────────────────────────────────────────────────────────


class TestAPIKeys:
    def test_list_api_keys(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"api_keys": [{"id": "k1", "name": "K", "organization_id": "o1"}]}
        )
        keys = MachineAuthClient(session=session).list_api_keys("o1")
        assert len(keys) == 1 and isinstance(keys[0], APIKey)

    def test_create_api_key(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"api_key": {"id": "k1"}, "key": "raw-key"}
        )
        assert MachineAuthClient(session=session).create_api_key("o1", "K")["key"] == "raw-key"

    def test_delete_api_key(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(status_code=204, text="")
        MachineAuthClient(session=session).delete_api_key("o1", "k1")
        session.request.assert_called_once()


# ── Webhooks ─────────────────────────────────────────────────────────


class TestWebhooks:
    def test_list_webhooks(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"webhooks": [{"id": "w1", "name": "W", "url": "http://x", "events": ["a"]}]}
        )
        hooks = MachineAuthClient(session=session).list_webhooks()
        assert len(hooks) == 1 and isinstance(hooks[0], WebhookConfig)

    def test_create_webhook(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(json_data={"webhook": {"id": "w1"}, "secret": "sec"})
        assert MachineAuthClient(session=session).create_webhook("W", "http://x", ["agent.created"])["secret"] == "sec"

    def test_get_webhook(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"webhook": {"id": "w1", "name": "W", "url": "http://x", "events": ["a"]}}
        )
        assert isinstance(MachineAuthClient(session=session).get_webhook("w1"), WebhookConfig)

    def test_update_webhook(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"webhook": {"id": "w1", "name": "Updated", "url": "http://x", "events": ["a"]}}
        )
        assert MachineAuthClient(session=session).update_webhook("w1", name="Updated").name == "Updated"

    def test_delete_webhook(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(status_code=204, text="")
        MachineAuthClient(session=session).delete_webhook("w1")

    def test_test_webhook(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(json_data={"success": True, "status_code": 200})
        assert MachineAuthClient(session=session).test_webhook("w1", "agent.created")["success"] is True

    def test_list_deliveries(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            json_data={"deliveries": [{"id": "d1", "webhook_config_id": "w1", "event": "a", "status": "delivered"}]}
        )
        deliveries = MachineAuthClient(session=session).list_deliveries("w1")
        assert len(deliveries) == 1 and isinstance(deliveries[0], WebhookDelivery)

    def test_list_webhook_events(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(json_data={"events": ["agent.created", "agent.deleted"]})
        assert MachineAuthClient(session=session).list_webhook_events() == ["agent.created", "agent.deleted"]


# ── Error handling ───────────────────────────────────────────────────


class TestErrorHandling:
    def test_api_error_raises(self) -> None:
        session = MagicMock()
        session.request.return_value = _mock_response(
            status_code=404, json_data={"message": "not found"}, reason="Not Found"
        )
        with pytest.raises(MachineAuthError) as exc_info:
            MachineAuthClient(session=session).get_agent("missing")
        assert exc_info.value.status_code == 404

    def test_auth_headers_with_token(self) -> None:
        client = MachineAuthClient(access_token="bearer-tok")
        assert client._auth_headers(None)["Authorization"] == "Bearer bearer-tok"

    def test_auth_headers_override(self) -> None:
        client = MachineAuthClient(access_token="default")
        assert client._auth_headers("override")["Authorization"] == "Bearer override"

    def test_auth_headers_empty(self) -> None:
        assert MachineAuthClient()._auth_headers(None) == {}
