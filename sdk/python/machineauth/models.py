from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional


@dataclass
class AutoRefreshOptions:
    enabled: bool
    refresh_buffer_seconds: int = 60


@dataclass
class Agent:
    id: str
    name: str
    client_id: str
    scopes: List[str] = field(default_factory=list)
    organization_id: Optional[str] = None
    team_id: Optional[str] = None
    public_key: Optional[str] = None
    is_active: bool = True
    created_at: Optional[str] = None
    updated_at: Optional[str] = None
    expires_at: Optional[str] = None
    token_count: int = 0
    refresh_count: int = 0
    last_activity_at: Optional[str] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Agent":
        return cls(
            id=data.get("id", ""),
            name=data.get("name", ""),
            client_id=data.get("client_id", ""),
            scopes=data.get("scopes") or [],
            organization_id=data.get("organization_id"),
            team_id=data.get("team_id"),
            public_key=data.get("public_key"),
            is_active=data.get("is_active", True),
            created_at=data.get("created_at"),
            updated_at=data.get("updated_at"),
            expires_at=data.get("expires_at"),
            token_count=data.get("token_count", 0),
            refresh_count=data.get("refresh_count", 0),
            last_activity_at=data.get("last_activity_at"),
        )


@dataclass
class TokenResponse:
    access_token: str
    token_type: str
    expires_in: Optional[int] = None
    scope: Optional[str] = None
    issued_at: Optional[str] = None
    refresh_token: Optional[str] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "TokenResponse":
        return cls(
            access_token=data.get("access_token", ""),
            token_type=data.get("token_type", ""),
            expires_in=data.get("expires_in"),
            scope=data.get("scope"),
            issued_at=data.get("issued_at"),
            refresh_token=data.get("refresh_token"),
        )


@dataclass
class AgentUsage:
    agent: Agent
    organization_id: Optional[str] = None
    token_count: int = 0
    refresh_count: int = 0
    rotation_history: List[Dict[str, Any]] = field(default_factory=list)

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "AgentUsage":
        agent_data = data.get("agent", data)
        return cls(
            agent=Agent.from_dict(agent_data) if isinstance(agent_data, dict) else agent_data,
            organization_id=data.get("organization_id"),
            token_count=data.get("token_count", 0),
            refresh_count=data.get("refresh_count", 0),
            rotation_history=data.get("rotation_history") or [],
        )


@dataclass
class Organization:
    id: str
    name: str
    slug: str
    owner_email: Optional[str] = None
    jwt_issuer: Optional[str] = None
    jwt_expiry_secs: Optional[int] = None
    allowed_origins: Optional[List[str]] = None
    created_at: Optional[str] = None
    updated_at: Optional[str] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Organization":
        return cls(
            id=data.get("id", ""),
            name=data.get("name", ""),
            slug=data.get("slug", ""),
            owner_email=data.get("owner_email"),
            jwt_issuer=data.get("jwt_issuer"),
            jwt_expiry_secs=data.get("jwt_expiry_secs"),
            allowed_origins=data.get("allowed_origins"),
            created_at=data.get("created_at"),
            updated_at=data.get("updated_at"),
        )


@dataclass
class Team:
    id: str
    name: str
    organization_id: str
    description: Optional[str] = None
    created_at: Optional[str] = None
    updated_at: Optional[str] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Team":
        return cls(
            id=data.get("id", ""),
            name=data.get("name", ""),
            organization_id=data.get("organization_id", ""),
            description=data.get("description"),
            created_at=data.get("created_at"),
            updated_at=data.get("updated_at"),
        )


@dataclass
class APIKey:
    id: str
    name: str
    organization_id: str
    key_prefix: Optional[str] = None
    team_id: Optional[str] = None
    is_active: bool = True
    created_at: Optional[str] = None
    expires_at: Optional[str] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "APIKey":
        return cls(
            id=data.get("id", ""),
            name=data.get("name", ""),
            organization_id=data.get("organization_id", ""),
            key_prefix=data.get("key_prefix"),
            team_id=data.get("team_id"),
            is_active=data.get("is_active", True),
            created_at=data.get("created_at"),
            expires_at=data.get("expires_at"),
        )


@dataclass
class WebhookConfig:
    id: str
    name: str
    url: str
    events: List[str]
    secret: Optional[str] = None
    organization_id: Optional[str] = None
    team_id: Optional[str] = None
    is_active: bool = True
    max_retries: int = 10
    retry_backoff_base: int = 2
    created_at: Optional[str] = None
    updated_at: Optional[str] = None
    last_tested_at: Optional[str] = None
    consecutive_fails: int = 0

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "WebhookConfig":
        return cls(
            id=data.get("id", ""),
            name=data.get("name", ""),
            url=data.get("url", ""),
            events=data.get("events") or [],
            secret=data.get("secret"),
            organization_id=data.get("organization_id"),
            team_id=data.get("team_id"),
            is_active=data.get("is_active", True),
            max_retries=data.get("max_retries", 10),
            retry_backoff_base=data.get("retry_backoff_base", 2),
            created_at=data.get("created_at"),
            updated_at=data.get("updated_at"),
            last_tested_at=data.get("last_tested_at"),
            consecutive_fails=data.get("consecutive_fails", 0),
        )


@dataclass
class WebhookDelivery:
    id: str
    webhook_config_id: str
    event: str
    status: str
    payload: Optional[str] = None
    headers: Optional[str] = None
    attempts: int = 0
    last_attempt_at: Optional[str] = None
    last_error: Optional[str] = None
    next_retry_at: Optional[str] = None
    created_at: Optional[str] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "WebhookDelivery":
        return cls(
            id=data.get("id", ""),
            webhook_config_id=data.get("webhook_config_id", ""),
            event=data.get("event", ""),
            status=data.get("status", ""),
            payload=data.get("payload"),
            headers=data.get("headers"),
            attempts=data.get("attempts", 0),
            last_attempt_at=data.get("last_attempt_at"),
            last_error=data.get("last_error"),
            next_retry_at=data.get("next_retry_at"),
            created_at=data.get("created_at"),
        )
