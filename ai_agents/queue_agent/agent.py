import os
import json
from dataclasses import dataclass
from typing import Any, Dict, List, Optional

import httpx
from google import genai
from google.genai import types
from google.adk.agents import Agent

# ----------------------------
# Config
# ----------------------------
GEMINI_MODEL = os.getenv("GEMINI_MODEL", "gemini-2.5-flash")
QUEUE_BASE_URL = os.environ["QUEUE_BASE_URL"]  # raise if missing

client = genai.Client()  # reads GEMINI_API_KEY from env

# single shared httpx client
_http = httpx.Client(base_url=QUEUE_BASE_URL, timeout=10.0)

# ----------------------------
# Tool functions (the agent will call these)
# ----------------------------

def change_priority_by_family(family: str, new_priority: int) -> dict:
    """Change priority for all ads in a given game family.

    Args:
        family: The 'GameFamily' name, e.g. "RPG-Fantasy".
        new_priority: Target priority (int).
    Returns:
        JSON dict of the API response (ok:true).
    """
    resp = _http.post("/reprioritize/family", json={
        "family": family, "newPriority": new_priority
    })
    resp.raise_for_status()
    return resp.json()

def reprioritize_by_age(age: str, new_priority: int) -> dict:
    """Set priority for all ads waiting older than a duration.

    Args:
        age: Go-style duration string, e.g. "10m", "5s", "1h30m".
        new_priority: Target priority.
    """
    resp = _http.post("/reprioritize/age", json={
        "age": age, "newPriority": new_priority
    })
    resp.raise_for_status()
    return resp.json()

def set_anti_starvation(enable: bool) -> dict:
    """Enable or disable anti-starvation behavior.

    Args:
        enable: True to enable anti-starvation; False to disable (aka 'starvation mode').
    """
    resp = _http.post("/settings/antiStarvation", json={"enable": enable})
    resp.raise_for_status()
    return resp.json()

def set_maximum_wait(seconds: int) -> dict:
    """Set global maximum wait time cap (seconds) for all ads."""
    resp = _http.post("/settings/maximumWait", json={"maximumWait": seconds})
    resp.raise_for_status()
    return resp.json()

def peek_next(n: int) -> List[Dict[str, Any]]:
    """Preview the next N ads in the order Dequeue would process them."""
    resp = _http.get("/peek", params={"n": n})
    resp.raise_for_status()
    return resp.json()

def list_waiting_longer_than(age: str) -> List[Dict[str, Any]]:
    """List ads waiting longer than a duration string (e.g., '5m')."""
    resp = _http.get("/waiting", params={"age": age})
    resp.raise_for_status()
    return resp.json()

def distribution_by_priority() -> Dict[str, Any]:
    """Get distribution of items by priority."""
    resp = _http.get("/distribution")
    resp.raise_for_status()
    return resp.json()

root_agent = Agent(
    name="priority_queue_agent",
    model="gemini-2.0-flash",
    description=(
        "You are a queue control agent for a video ad processing system. You can call tools to modify the queue and fetch analytics"
    ),
    instruction=(
        """You are a queue control agent for a video ad processing system.
You can call tools to modify the queue and fetch analytics.
Interpret user commands precisely. If a value looks like a duration, use Go-style
durations (e.g., '10m', '5s', '1h30m').
Map phrasing:
- "Enable starvation mode" => disable anti-starvation (set_anti_starvation(enable=false)).
- "Disable starvation mode" OR "Enable anti-starvation" => set_anti_starvation(true).
Confirm your actions with concise summaries and include key numbers or lists when helpful.
If a number of items is requested (e.g., 'next 5'), call peek_next with that N."""
    ),
    tools=[change_priority_by_family,
    reprioritize_by_age,
    set_anti_starvation,
    set_maximum_wait,
    peek_next,
    list_waiting_longer_than,
    distribution_by_priority,
],
)
