"""
Prompt management for fetching and compiling prompts.
"""

from __future__ import annotations

import re
import time
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Union

from agenttrace.context import get_client


@dataclass
class PromptVersion:
    """Represents a specific version of a prompt."""

    id: str
    version: int
    prompt: str
    config: Dict[str, Any] = field(default_factory=dict)
    labels: List[str] = field(default_factory=list)
    created_at: str = ""

    def compile(self, **variables: Any) -> str:
        """
        Compile the prompt with variables.

        Args:
            **variables: Variable values to substitute

        Returns:
            Compiled prompt string

        Example:
            prompt.compile(name="Alice", topic="Python")
        """
        result = self.prompt
        for key, value in variables.items():
            # Support both {{var}} and {var} syntax
            result = result.replace(f"{{{{{key}}}}}", str(value))
            result = result.replace(f"{{{key}}}", str(value))
        return result

    def compile_chat(
        self, **variables: Any
    ) -> List[Dict[str, str]]:
        """
        Compile as chat messages (if prompt is in chat format).

        Expects prompt to be in the format:
        system: You are a helpful assistant.
        user: Hello, {{name}}!
        assistant: Hi {{name}}, how can I help?

        Args:
            **variables: Variable values to substitute

        Returns:
            List of message dicts with 'role' and 'content' keys
        """
        compiled = self.compile(**variables)
        messages = []

        # Parse role: content format
        lines = compiled.strip().split("\n")
        current_role = None
        current_content = []

        for line in lines:
            # Check for role prefix
            role_match = re.match(r"^(system|user|assistant|function):\s*(.*)$", line, re.IGNORECASE)
            if role_match:
                # Save previous message
                if current_role and current_content:
                    messages.append({
                        "role": current_role.lower(),
                        "content": "\n".join(current_content).strip()
                    })
                current_role = role_match.group(1)
                current_content = [role_match.group(2)] if role_match.group(2) else []
            else:
                current_content.append(line)

        # Don't forget the last message
        if current_role and current_content:
            messages.append({
                "role": current_role.lower(),
                "content": "\n".join(current_content).strip()
            })

        return messages

    def get_variables(self) -> List[str]:
        """
        Extract variable names from the prompt.

        Returns:
            List of variable names found in the prompt
        """
        # Find both {{var}} and {var} patterns
        double_brace = re.findall(r"\{\{(\w+)\}\}", self.prompt)
        single_brace = re.findall(r"\{(\w+)\}", self.prompt)

        # Combine and deduplicate
        return list(set(double_brace + single_brace))


class Prompt:
    """
    Main prompt class for fetching and managing prompts.

    Example:
        # Get a prompt by name
        prompt = Prompt.get("my-prompt")
        compiled = prompt.compile(name="Alice")

        # Get a specific version
        prompt = Prompt.get("my-prompt", version=2)

        # Get by label
        prompt = Prompt.get("my-prompt", label="production")
    """

    # In-memory cache for prompts
    _cache: Dict[str, tuple[PromptVersion, float]] = {}
    _cache_ttl: float = 60.0  # 1 minute default TTL

    def __init__(
        self,
        name: str,
        version: Optional[int] = None,
        label: Optional[str] = None,
        fallback: Optional[str] = None,
        cache_ttl: Optional[float] = None,
    ):
        """
        Initialize a Prompt instance.

        Args:
            name: Prompt name
            version: Specific version number
            label: Label to fetch (e.g., "production", "latest")
            fallback: Fallback prompt if fetch fails
            cache_ttl: Cache TTL in seconds (None to use default)
        """
        self.name = name
        self.version = version
        self.label = label
        self.fallback = fallback
        self._cache_ttl = cache_ttl if cache_ttl is not None else Prompt._cache_ttl
        self._prompt_version: Optional[PromptVersion] = None

    @classmethod
    def get(
        cls,
        name: str,
        version: Optional[int] = None,
        label: Optional[str] = None,
        fallback: Optional[str] = None,
        cache_ttl: Optional[float] = None,
    ) -> PromptVersion:
        """
        Fetch a prompt from the server.

        Args:
            name: Prompt name
            version: Specific version number
            label: Label to fetch (e.g., "production")
            fallback: Fallback prompt if fetch fails
            cache_ttl: Cache TTL in seconds

        Returns:
            PromptVersion object

        Raises:
            ValueError: If prompt not found and no fallback provided
        """
        instance = cls(
            name=name,
            version=version,
            label=label,
            fallback=fallback,
            cache_ttl=cache_ttl,
        )
        return instance.fetch()

    def fetch(self) -> PromptVersion:
        """
        Fetch the prompt from the server.

        Returns:
            PromptVersion object
        """
        cache_key = self._get_cache_key()

        # Check cache first
        if cache_key in Prompt._cache:
            cached, cached_at = Prompt._cache[cache_key]
            if time.time() - cached_at < self._cache_ttl:
                return cached

        client = get_client()
        if client is None:
            return self._get_fallback()

        try:
            # Make API request
            params: Dict[str, Any] = {"name": self.name}
            if self.version is not None:
                params["version"] = self.version
            if self.label is not None:
                params["label"] = self.label

            response = client._transport.get("/api/public/prompts", params=params)

            if response and "id" in response:
                prompt_version = PromptVersion(
                    id=response["id"],
                    version=response.get("version", 1),
                    prompt=response.get("prompt", ""),
                    config=response.get("config", {}),
                    labels=response.get("labels", []),
                    created_at=response.get("createdAt", ""),
                )

                # Update cache
                Prompt._cache[cache_key] = (prompt_version, time.time())
                self._prompt_version = prompt_version
                return prompt_version

            return self._get_fallback()

        except Exception as e:
            # Log error and use fallback
            return self._get_fallback()

    def _get_cache_key(self) -> str:
        """Generate cache key for this prompt."""
        parts = [self.name]
        if self.version is not None:
            parts.append(f"v{self.version}")
        if self.label is not None:
            parts.append(f"l:{self.label}")
        return ":".join(parts)

    def _get_fallback(self) -> PromptVersion:
        """Get fallback prompt version."""
        if self.fallback is not None:
            return PromptVersion(
                id="fallback",
                version=0,
                prompt=self.fallback,
                labels=["fallback"],
            )
        raise ValueError(f"Prompt '{self.name}' not found and no fallback provided")

    def compile(self, **variables: Any) -> str:
        """
        Fetch and compile the prompt with variables.

        Args:
            **variables: Variable values to substitute

        Returns:
            Compiled prompt string
        """
        prompt_version = self.fetch()
        return prompt_version.compile(**variables)

    def compile_chat(self, **variables: Any) -> List[Dict[str, str]]:
        """
        Fetch and compile as chat messages.

        Args:
            **variables: Variable values to substitute

        Returns:
            List of message dicts
        """
        prompt_version = self.fetch()
        return prompt_version.compile_chat(**variables)

    @classmethod
    def set_cache_ttl(cls, ttl: float) -> None:
        """
        Set the default cache TTL for all prompts.

        Args:
            ttl: TTL in seconds
        """
        cls._cache_ttl = ttl

    @classmethod
    def clear_cache(cls) -> None:
        """Clear the prompt cache."""
        cls._cache.clear()

    @classmethod
    def invalidate(cls, name: str) -> None:
        """
        Invalidate cache for a specific prompt.

        Args:
            name: Prompt name to invalidate
        """
        keys_to_remove = [k for k in cls._cache if k.startswith(name)]
        for key in keys_to_remove:
            del cls._cache[key]


def get_prompt(
    name: str,
    version: Optional[int] = None,
    label: Optional[str] = None,
    fallback: Optional[str] = None,
) -> PromptVersion:
    """
    Convenience function to fetch a prompt.

    Args:
        name: Prompt name
        version: Specific version number
        label: Label to fetch
        fallback: Fallback prompt if fetch fails

    Returns:
        PromptVersion object
    """
    return Prompt.get(name=name, version=version, label=label, fallback=fallback)
