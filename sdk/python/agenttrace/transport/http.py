"""
HTTP transport for sending data to AgentTrace.
"""

from __future__ import annotations

import json
import logging
import time
from typing import Any, Dict, List, Optional

import httpx

logger = logging.getLogger("agenttrace")


class HttpTransport:
    """
    HTTP transport for communicating with the AgentTrace API.

    Handles retries, timeouts, and error handling.
    """

    def __init__(
        self,
        host: str,
        api_key: str,
        timeout: float = 10.0,
        max_retries: int = 3,
        public_key: Optional[str] = None,
    ):
        """
        Initialize the HTTP transport.

        Args:
            host: AgentTrace API host URL
            api_key: API key for authentication
            timeout: Request timeout in seconds
            max_retries: Maximum number of retries for failed requests
            public_key: Optional public key
        """
        self.host = host.rstrip("/")
        self.api_key = api_key
        self.public_key = public_key
        self.timeout = timeout
        self.max_retries = max_retries

        # Create HTTP client with connection pooling
        self._client = httpx.Client(
            timeout=httpx.Timeout(timeout),
            limits=httpx.Limits(max_keepalive_connections=10, max_connections=20),
        )

    def _get_headers(self) -> Dict[str, str]:
        """Get request headers with authentication."""
        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self.api_key}",
            "User-Agent": "agenttrace-python/0.1.0",
        }
        if self.public_key:
            headers["X-Langfuse-Public-Key"] = self.public_key
        return headers

    def send_batch(self, events: List[Dict[str, Any]]) -> bool:
        """
        Send a batch of events to the AgentTrace API.

        Args:
            events: List of event dictionaries

        Returns:
            True if successful, False otherwise
        """
        if not events:
            return True

        url = f"{self.host}/api/public/ingestion"
        payload = {"batch": events}

        for attempt in range(self.max_retries):
            try:
                response = self._client.post(
                    url,
                    headers=self._get_headers(),
                    json=payload,
                )

                if response.status_code in (200, 201, 207):
                    # Log any partial failures in 207 response
                    if response.status_code == 207:
                        try:
                            result = response.json()
                            if "errors" in result:
                                for error in result["errors"]:
                                    logger.warning(f"Partial failure: {error}")
                        except Exception:
                            pass
                    return True

                if response.status_code == 429:
                    # Rate limited - wait and retry
                    retry_after = int(response.headers.get("Retry-After", 5))
                    logger.warning(f"Rate limited, waiting {retry_after}s")
                    time.sleep(retry_after)
                    continue

                if response.status_code >= 500:
                    # Server error - retry with backoff
                    wait_time = (2**attempt) * 0.5
                    logger.warning(
                        f"Server error {response.status_code}, "
                        f"retrying in {wait_time}s (attempt {attempt + 1}/{self.max_retries})"
                    )
                    time.sleep(wait_time)
                    continue

                # Client error - don't retry
                logger.error(
                    f"Client error {response.status_code}: {response.text}"
                )
                return False

            except httpx.TimeoutException:
                wait_time = (2**attempt) * 0.5
                logger.warning(
                    f"Request timeout, retrying in {wait_time}s "
                    f"(attempt {attempt + 1}/{self.max_retries})"
                )
                time.sleep(wait_time)

            except httpx.RequestError as e:
                wait_time = (2**attempt) * 0.5
                logger.warning(
                    f"Request error: {e}, retrying in {wait_time}s "
                    f"(attempt {attempt + 1}/{self.max_retries})"
                )
                time.sleep(wait_time)

            except Exception as e:
                logger.error(f"Unexpected error sending batch: {e}")
                return False

        logger.error(f"Failed to send batch after {self.max_retries} attempts")
        return False

    def get(
        self,
        path: str,
        params: Optional[Dict[str, Any]] = None,
    ) -> Optional[Dict[str, Any]]:
        """
        Make a GET request to the API.

        Args:
            path: API path
            params: Query parameters

        Returns:
            Response data or None on error
        """
        url = f"{self.host}{path}"

        for attempt in range(self.max_retries):
            try:
                response = self._client.get(
                    url,
                    headers=self._get_headers(),
                    params=params,
                )

                if response.status_code == 200:
                    return response.json()

                if response.status_code == 404:
                    return None

                if response.status_code == 429:
                    retry_after = int(response.headers.get("Retry-After", 5))
                    time.sleep(retry_after)
                    continue

                if response.status_code >= 500:
                    wait_time = (2**attempt) * 0.5
                    time.sleep(wait_time)
                    continue

                logger.error(f"GET {path} failed: {response.status_code}")
                return None

            except httpx.TimeoutException:
                wait_time = (2**attempt) * 0.5
                time.sleep(wait_time)

            except httpx.RequestError as e:
                logger.warning(f"Request error on GET {path}: {e}")
                wait_time = (2**attempt) * 0.5
                time.sleep(wait_time)

            except Exception as e:
                logger.error(f"Unexpected error on GET {path}: {e}")
                return None

        return None

    def post(
        self,
        path: str,
        data: Dict[str, Any],
    ) -> Optional[Dict[str, Any]]:
        """
        Make a POST request to the API.

        Args:
            path: API path
            data: Request body

        Returns:
            Response data or None on error
        """
        url = f"{self.host}{path}"

        for attempt in range(self.max_retries):
            try:
                response = self._client.post(
                    url,
                    headers=self._get_headers(),
                    json=data,
                )

                if response.status_code in (200, 201):
                    return response.json()

                if response.status_code == 429:
                    retry_after = int(response.headers.get("Retry-After", 5))
                    time.sleep(retry_after)
                    continue

                if response.status_code >= 500:
                    wait_time = (2**attempt) * 0.5
                    time.sleep(wait_time)
                    continue

                logger.error(f"POST {path} failed: {response.status_code}")
                return None

            except httpx.TimeoutException:
                wait_time = (2**attempt) * 0.5
                time.sleep(wait_time)

            except httpx.RequestError as e:
                logger.warning(f"Request error on POST {path}: {e}")
                wait_time = (2**attempt) * 0.5
                time.sleep(wait_time)

            except Exception as e:
                logger.error(f"Unexpected error on POST {path}: {e}")
                return None

        return None

    def close(self) -> None:
        """Close the HTTP client."""
        self._client.close()

    def __del__(self):
        """Cleanup on destruction."""
        try:
            self._client.close()
        except Exception:
            pass


class AsyncHttpTransport:
    """
    Async HTTP transport for communicating with the AgentTrace API.
    """

    def __init__(
        self,
        host: str,
        api_key: str,
        timeout: float = 10.0,
        max_retries: int = 3,
        public_key: Optional[str] = None,
    ):
        """Initialize the async HTTP transport."""
        self.host = host.rstrip("/")
        self.api_key = api_key
        self.public_key = public_key
        self.timeout = timeout
        self.max_retries = max_retries
        self._client: Optional[httpx.AsyncClient] = None

    async def _get_client(self) -> httpx.AsyncClient:
        """Get or create the async HTTP client."""
        if self._client is None:
            self._client = httpx.AsyncClient(
                timeout=httpx.Timeout(self.timeout),
                limits=httpx.Limits(max_keepalive_connections=10, max_connections=20),
            )
        return self._client

    def _get_headers(self) -> Dict[str, str]:
        """Get request headers with authentication."""
        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self.api_key}",
            "User-Agent": "agenttrace-python/0.1.0",
        }
        if self.public_key:
            headers["X-Langfuse-Public-Key"] = self.public_key
        return headers

    async def send_batch(self, events: List[Dict[str, Any]]) -> bool:
        """Send a batch of events asynchronously."""
        if not events:
            return True

        client = await self._get_client()
        url = f"{self.host}/api/public/ingestion"
        payload = {"batch": events}

        for attempt in range(self.max_retries):
            try:
                response = await client.post(
                    url,
                    headers=self._get_headers(),
                    json=payload,
                )

                if response.status_code in (200, 201, 207):
                    return True

                if response.status_code == 429:
                    retry_after = int(response.headers.get("Retry-After", 5))
                    await self._async_sleep(retry_after)
                    continue

                if response.status_code >= 500:
                    wait_time = (2**attempt) * 0.5
                    await self._async_sleep(wait_time)
                    continue

                return False

            except Exception as e:
                logger.warning(f"Async request error: {e}")
                wait_time = (2**attempt) * 0.5
                await self._async_sleep(wait_time)

        return False

    async def _async_sleep(self, seconds: float) -> None:
        """Async sleep helper."""
        import asyncio
        await asyncio.sleep(seconds)

    async def close(self) -> None:
        """Close the async HTTP client."""
        if self._client:
            await self._client.aclose()
            self._client = None
