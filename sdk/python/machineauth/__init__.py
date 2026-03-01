from .client import MachineAuthClient, MachineAuthError
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

# Async client imported lazily to avoid requiring httpx at import time
def __getattr__(name: str):
    if name == "AsyncMachineAuthClient":
        from .async_client import AsyncMachineAuthClient
        return AsyncMachineAuthClient
    raise AttributeError(f"module {__name__!r} has no attribute {name!r}")

__all__ = [
    "MachineAuthClient",
    "MachineAuthError",
    "AsyncMachineAuthClient",
    "Agent",
    "AgentUsage",
    "APIKey",
    "Organization",
    "Team",
    "TokenResponse",
    "WebhookConfig",
    "WebhookDelivery",
]
